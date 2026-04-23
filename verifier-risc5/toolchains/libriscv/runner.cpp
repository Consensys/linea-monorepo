#include <libriscv/machine.hpp>

#include <cstdlib>
#include <cstdint>
#include <fstream>
#include <iostream>
#include <iterator>
#include <stdexcept>
#include <string>
#include <vector>

namespace {

using Machine = riscv::Machine<riscv::RISCV64>;
using Address = Machine::address_t;

constexpr Address kGuestStatusBase = 0x80eff000ULL;
constexpr Address kGuestInputBase = 0x80f00000ULL;
constexpr Address kQEMUTestBase = 0x00100000ULL;
constexpr uint32_t kStatusMagic = 0x56535441U;
constexpr uint32_t kStatusVersion = 1U;
constexpr uint32_t kStatusCodeSuccess = 1U;
constexpr uint32_t kStatusCodeInputError = 2U;
constexpr uint32_t kStatusCodeMismatch = 3U;
constexpr uint32_t kQEMUTestPass = 0x5555U;
constexpr uint32_t kQEMUTestFail = 0x3333U;
constexpr uint64_t kDefaultMaxInstructions = 50'000'000ULL;

struct GuestStatus {
	uint32_t magic;
	uint32_t version;
	uint32_t code;
	uint32_t reserved;
	uint64_t result;
	uint64_t expected;
};

struct GuestFinish {
	bool seen = false;
	uint32_t value = 0;
};

std::vector<uint8_t> read_binary(const char* path) {
	std::ifstream stream(path, std::ios::binary);
	if (!stream) {
		throw std::runtime_error(std::string("failed to open ELF: ") + path);
	}

	return std::vector<uint8_t>(
		(std::istreambuf_iterator<char>(stream)),
		std::istreambuf_iterator<char>());
}

void write_trapped_value(riscv::Page& page, uint32_t offset, int mode, int64_t value) {
	switch (riscv::Page::trap_size(mode)) {
	case 1:
		page.page().template aligned_write<uint8_t>(offset, static_cast<uint8_t>(value));
		return;
	case 2:
		page.page().template aligned_write<uint16_t>(offset, static_cast<uint16_t>(value));
		return;
	case 4:
		page.page().template aligned_write<uint32_t>(offset, static_cast<uint32_t>(value));
		return;
	case 8:
		page.page().template aligned_write<uint64_t>(offset, static_cast<uint64_t>(value));
		return;
	default:
		return;
	}
}

bool read_guest_status(const Machine& machine, GuestStatus& status) {
	try {
		machine.copy_from_guest(&status, kGuestStatusBase, sizeof(status));
		return status.magic == kStatusMagic && status.version == kStatusVersion;
	} catch (...) {
		return false;
	}
}

int report_guest_status(const GuestStatus& status) {
	switch (status.code) {
	case kStatusCodeSuccess:
		std::cout << "libriscv runner: guest reported success\n";
		return 0;
	case kStatusCodeInputError:
		std::cerr << "libriscv runner: guest reported invalid input\n";
		return 5;
	case kStatusCodeMismatch:
		std::cerr << "libriscv runner: guest reported mismatch (got=0x"
			  << std::hex << status.result
			  << ", expected=0x" << status.expected
			  << std::dec << ")\n";
		return 4;
	default:
		std::cerr << "libriscv runner: guest reported unknown status code "
			  << status.code << '\n';
		return 6;
	}
}

uint64_t parse_max_instructions(int argc, char** argv) {
	if (argc < 4) {
		return kDefaultMaxInstructions;
	}

	char* end = nullptr;
	const auto parsed = std::strtoull(argv[3], &end, 0);
	if (end == nullptr || *end != '\0') {
		throw std::runtime_error(std::string("invalid instruction limit: ") + argv[3]);
	}

	return parsed;
}

std::string describe_guest_state(const Machine& machine) {
	try {
		return machine.cpu.current_instruction_to_string();
	} catch (...) {
		return "instruction unavailable";
	}
}

} // namespace

int main(int argc, char** argv) {
	if (argc < 3) {
		std::cerr << "usage: " << argv[0] << " <guest.elf> <input.bin> [max-instructions]\n";
		return 64;
	}

	try {
		const auto binary = read_binary(argv[1]);
		const auto input = read_binary(argv[2]);
		const auto max_instructions = parse_max_instructions(argc, argv);

		Machine::on_unhandled_csr = [] (Machine&, int csr, int, int) {
			throw riscv::MachineException(
				riscv::ILLEGAL_OPERATION,
				"unsupported CSR in bare-metal libriscv runner",
				static_cast<uint64_t>(csr));
		};

		riscv::MachineOptions<riscv::RISCV64> options {
			.memory_max = 64ULL << 20,
			.allow_write_exec_segment = true,
			.use_memory_arena = false,
		};
		Machine machine { binary, options };
		machine.copy_to_guest(kGuestInputBase, input.data(), input.size());

		GuestFinish finisher;
		auto& finisher_page =
			machine.memory.create_writable_pageno(riscv::Memory<riscv::RISCV64>::page_number(kQEMUTestBase));
		finisher_page.set_trap(
			[&] (riscv::Page& page, uint32_t offset, int mode, int64_t value) {
				switch (riscv::Page::trap_mode(mode)) {
				case riscv::TRAP_WRITE:
					write_trapped_value(page, offset, mode, value);
					if (offset == 0 && riscv::Page::trap_size(mode) >= 4) {
						finisher.seen = true;
						finisher.value = static_cast<uint32_t>(value);
						machine.stop();
					}
					return;
				default:
					return;
				}
			});

		try {
			const bool stopped_normally = machine.simulate<false>(max_instructions);

			GuestStatus status {};
			if (read_guest_status(machine, status)) {
				return report_guest_status(status);
			}

			if (!stopped_normally && machine.instruction_limit_reached()) {
				std::cerr << "libriscv runner: instruction limit reached before guest status was written\n";
				return 2;
			}

			if (finisher.seen) {
				std::cerr << "libriscv runner: guest stopped via finisher without a valid status page"
					  << " (value=0x" << std::hex << finisher.value << std::dec << ")\n";
				return 6;
			}

			std::cerr << "libriscv runner: guest stopped without producing a valid status page\n";
			return 3;
		} catch (const riscv::MachineException& err) {
			std::cerr << "libriscv runner: " << err.what()
				  << " (type=" << err.type()
				  << ", data=0x" << std::hex << err.data() << std::dec
				  << ", pc=0x" << std::hex << machine.cpu.pc() << std::dec << ")\n";
			std::cerr << "libriscv runner: " << describe_guest_state(machine) << '\n';
			return 1;
		} catch (const std::exception& err) {
			std::cerr << "libriscv runner: " << err.what()
				  << " (pc=0x" << std::hex << machine.cpu.pc() << std::dec << ")\n";
			std::cerr << "libriscv runner: " << describe_guest_state(machine) << '\n';
			return 1;
		}
	} catch (const std::exception& err) {
		std::cerr << "libriscv runner: " << err.what() << '\n';
		return 1;
	}
}
