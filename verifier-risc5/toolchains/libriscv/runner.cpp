#include <libriscv/machine.hpp>

#include <cstdlib>
#include <fstream>
#include <iostream>
#include <iterator>
#include <string>
#include <vector>

namespace {

using Machine = riscv::Machine<riscv::RISCV64>;
using Address = Machine::address_t;

constexpr Address kUARTBase = 0x10000000ULL;
constexpr uint32_t kUARTTHROffset = 0;
constexpr uint32_t kUARTLSROffset = 5;
constexpr uint8_t kUARTLSRTHRE = 0x20;
constexpr uint64_t kDefaultMaxInstructions = 50'000'000ULL;

std::vector<uint8_t> read_binary(const char* path) {
	std::ifstream stream(path, std::ios::binary);
	if (!stream) {
		throw std::runtime_error(std::string("failed to open ELF: ") + path);
	}

	return std::vector<uint8_t>(
		(std::istreambuf_iterator<char>(stream)),
		std::istreambuf_iterator<char>());
}

bool has_complete_result_line(const std::string& output) {
	const auto start = output.find("verifier result ");
	if (start == std::string::npos) {
		return false;
	}

	return output.find('\n', start) != std::string::npos;
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

uint64_t parse_max_instructions(int argc, char** argv) {
	if (argc < 3) {
		return kDefaultMaxInstructions;
	}

	char* end = nullptr;
	const auto parsed = std::strtoull(argv[2], &end, 0);
	if (end == nullptr || *end != '\0') {
		throw std::runtime_error(std::string("invalid instruction limit: ") + argv[2]);
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
	if (argc < 2) {
		std::cerr << "usage: " << argv[0] << " <guest.elf> [max-instructions]\n";
		return 64;
	}

	try {
		const auto binary = read_binary(argv[1]);
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

		std::string output;
		auto& uart_page =
			machine.memory.create_writable_pageno(riscv::Memory<riscv::RISCV64>::page_number(kUARTBase));
		uart_page.page().template aligned_write<uint8_t>(kUARTLSROffset, kUARTLSRTHRE);
		uart_page.set_trap(
			[&] (riscv::Page& page, uint32_t offset, int mode, int64_t value) {
				switch (riscv::Page::trap_mode(mode)) {
				case riscv::TRAP_READ:
					if (offset == kUARTLSROffset) {
						page.page().template aligned_write<uint8_t>(offset, kUARTLSRTHRE);
					}
					return;
				case riscv::TRAP_WRITE:
					if (offset == kUARTTHROffset && riscv::Page::trap_size(mode) >= 1) {
						const char ch = static_cast<char>(value & 0xff);
						output.push_back(ch);
						std::cout.put(ch);
						std::cout.flush();
						if (has_complete_result_line(output)) {
							machine.stop();
						}
					}
					write_trapped_value(page, offset, mode, value);
					return;
				default:
					return;
				}
			});

		try {
			const bool stopped_normally = machine.simulate<false>(max_instructions);
			if (has_complete_result_line(output)) {
				return 0;
			}

			if (!stopped_normally && machine.instruction_limit_reached()) {
				std::cerr << "libriscv runner: instruction limit reached before complete output\n";
				return 2;
			}

			std::cerr << "libriscv runner: guest stopped without producing a complete result line\n";
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
