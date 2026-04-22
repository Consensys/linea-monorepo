typedef unsigned char uint8_t;
typedef unsigned int uint32_t;
typedef unsigned long long uint64_t;

enum {
	GUEST_STATUS_BASE = 0x80eff000u,
	GUEST_STATUS_MAGIC = 0x56535441u,
	GUEST_STATUS_VERSION = 1u,
	GUEST_INPUT_BASE = 0x80f00000u,
	GUEST_INPUT_SIZE = 0x00100000u,
	GUEST_INPUT_MAGIC = 0x56524659u,
	GUEST_INPUT_VERSION = 1u,
	GUEST_INPUT_WORDS_OFFSET = 24u,
	STATUS_CODE_SUCCESS = 1u,
	STATUS_CODE_INPUT_ERROR = 2u,
	STATUS_CODE_MISMATCH = 3u,
	QEMU_TEST_BASE = 0x00100000u,
	QEMU_TEST_PASS = 0x5555u,
	QEMU_TEST_FAIL = 0x3333u,
};

typedef struct guest_input_header {
	uint32_t magic;
	uint32_t version;
	uint32_t word_count;
	uint32_t reserved;
	uint64_t expected;
} guest_input_header_t;

typedef struct guest_status {
	uint32_t magic;
	uint32_t version;
	uint32_t code;
	uint32_t reserved;
	uint64_t result;
	uint64_t expected;
} guest_status_t;

static volatile guest_input_header_t *const input_header = (volatile guest_input_header_t *)GUEST_INPUT_BASE;
static volatile uint64_t *const input_words = (volatile uint64_t *)(GUEST_INPUT_BASE + GUEST_INPUT_WORDS_OFFSET);
static volatile guest_status_t *const guest_status = (volatile guest_status_t *)GUEST_STATUS_BASE;
static volatile uint32_t *const qemu_test = (volatile uint32_t *)QEMU_TEST_BASE;

static void halt_forever(void) {
	for (;;) {
	}
}

static void report_status(uint32_t code, uint64_t result, uint64_t expected) {
	guest_status->magic = GUEST_STATUS_MAGIC;
	guest_status->version = GUEST_STATUS_VERSION;
	guest_status->code = code;
	guest_status->reserved = 0u;
	guest_status->result = result;
	guest_status->expected = expected;

	uint32_t finisher_value = QEMU_TEST_PASS;
	if (code != STATUS_CODE_SUCCESS) {
		finisher_value = (code << 16) | QEMU_TEST_FAIL;
	}

	*qemu_test = finisher_value;
	halt_forever();
}

static uint64_t mix64(uint64_t value) {
	value ^= value >> 30;
	value *= 0xbf58476d1ce4e5b9ull;
	value ^= value >> 27;
	value *= 0x94d049bb133111ebull;
	value ^= value >> 31;
	return value;
}

static uint64_t compute(const volatile uint64_t *words, uint32_t word_count) {
	uint64_t acc = 0x9e3779b97f4a7c15ull;

	for (uint32_t i = 0; i < word_count; i++) {
		acc = mix64(acc ^ words[i]);
	}

	return acc;
}

void guest_main(void) {
	const uint32_t max_words = (GUEST_INPUT_SIZE - GUEST_INPUT_WORDS_OFFSET) / 8u;
	if (input_header->magic != GUEST_INPUT_MAGIC || input_header->version != GUEST_INPUT_VERSION ||
	    input_header->word_count > max_words) {
		report_status(STATUS_CODE_INPUT_ERROR, 0u, 0u);
	}

	const uint64_t result = compute(input_words, input_header->word_count);
	if (result == input_header->expected) {
		report_status(STATUS_CODE_SUCCESS, result, result);
	}

	report_status(STATUS_CODE_MISMATCH, result, input_header->expected);
}
