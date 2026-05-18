SHELL := /bin/bash

THIS_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
EXAMPLES_MAKEFILE := $(abspath $(THIS_DIR)/../../..)/Makefile
VECTORS := $(THIS_DIR)blake2f.all
RUST_TEST := blake/blake_with_in_bytes.rs
SELECTOR ?=

.PHONY: all
all:
	@set -euo pipefail; \
	selector="$(SELECTOR)"; \
	cases=$$(awk -v selector="$$selector" ' \
		BEGIN { \
			if (selector != "" && selector !~ /^[0-9]+(-[0-9]+)?$$/) { \
				print "SELECTOR must be N or START-END" > "/dev/stderr"; \
				exit 2; \
			} \
			if (selector ~ /-/) split(selector, bounds, "-"); \
		} \
		selector == "" { print NR ":" $$0; next } \
		selector ~ /^[0-9]+$$/ && NR == selector { print NR ":" $$0; next } \
		selector ~ /^[0-9]+-[0-9]+$$/ && NR >= bounds[1] && NR <= bounds[2] { print NR ":" $$0 } \
	' "$(VECTORS)") || exit 1; \
	[[ -n "$$cases" ]] || { echo "no cases matched: $$selector"; exit 1; }; \
	failures=""; \
	while IFS=: read -r n in_bytes; do \
		printf 'case %s... ' "$$n"; \
		output=$$(make --no-print-directory -f "$(EXAMPLES_MAKEFILE)" TEST="$(RUST_TEST)" IN_BYTES="$$in_bytes" 2>&1) || true; \
		result_line=$$(printf '%s\n' "$$output" | sed '/^[[:space:]]*$$/d' | tail -n 2 | head -n 1); \
		if [[ "$$result_line" == "Program exited successfully (exit with code 0)." ]]; then \
			echo ok; \
		else \
			echo "failed ($$result_line)"; \
			failures+="$$n: $$result_line"$$'\n'; \
		fi; \
	done <<< "$$cases"; \
	if [[ -n "$$failures" ]]; then \
		echo "failed cases:"; \
		printf '%s' "$$failures"; \
		exit 1; \
	fi; \
	echo "all cases passed"
