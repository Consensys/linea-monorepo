typedef unsigned char uint8_t;
typedef unsigned long long uint64_t;

enum {
	UART_BASE = 0x10000000u,
	UART_THR = 0u,
	UART_LSR = 5u,
	UART_LSR_THRE = 0x20u,
};

static volatile uint8_t *const uart = (volatile uint8_t *)UART_BASE;

static void uart_write_byte(uint8_t value) {
	while ((uart[UART_LSR] & UART_LSR_THRE) == 0) {
	}

	uart[UART_THR] = value;
}

static void uart_write_string(const char *value) {
	while (*value != '\0') {
		uart_write_byte((uint8_t)*value);
		value++;
	}
}

static void uart_write_hex64(uint64_t value) {
	static const char hex[] = "0123456789abcdef";

	for (int shift = 60; shift >= 0; shift -= 4) {
		uart_write_byte((uint8_t)hex[(value >> shift) & 0x0fu]);
	}
}

static uint64_t mix64(uint64_t value) {
	value ^= value >> 30;
	value *= 0xbf58476d1ce4e5b9ull;
	value ^= value >> 27;
	value *= 0x94d049bb133111ebull;
	value ^= value >> 31;
	return value;
}

static uint64_t compute(void) {
	static const uint64_t input_words[] = {
		0x0123456789abcdefull,
		0xfedcba9876543210ull,
		0x0f1e2d3c4b5a6978ull,
		0x8877665544332211ull,
		0x13579bdf2468ace0ull,
		0xc001d00dcafef00dull,
	};

	uint64_t acc = 0x9e3779b97f4a7c15ull;

	for (unsigned long i = 0; i < sizeof(input_words) / sizeof(input_words[0]); i++) {
		acc = mix64(acc ^ input_words[i]);
	}

	return acc;
}

void guest_main(void) {
	uint64_t result = compute();

	uart_write_string("verifier result 0x");
	uart_write_hex64(result);
	uart_write_string("\r\n");

	for (;;) {
	}
}
