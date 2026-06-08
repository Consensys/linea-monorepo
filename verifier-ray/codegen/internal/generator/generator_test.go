package generator_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/verifier-ray/codegen/internal/generator"
)

func TestGenerate_ZeroRounds(t *testing.T) {
	var buf bytes.Buffer
	err := generator.Generate(generator.System{}, &buf)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	src := buf.String()
	if strings.Contains(src, "// Round ") {
		t.Fatalf("zero-round fallback must not emit round comments:\n%s", src)
	}
	if !strings.Contains(src, ".round_coin_counts = &[_]usize{ 0 },") {
		t.Fatalf("zero-round fallback must emit a real zero-round spec:\n%s", src)
	}
	if !strings.Contains(src, ".round_coin_offsets = &[_]usize{ 0 },") {
		t.Fatalf("zero-round fallback must emit a real zero-round offset schedule:\n%s", src)
	}
}

func TestGenerate_EmitsSpecAndSystems(t *testing.T) {
	var buf bytes.Buffer
	err := generator.Generate(generator.System{
		Rounds: []generator.Round{
			{ID: 0, VerifierActions: []string{"commitment"}},
			{ID: 1, VerifierActions: []string{"quotient"}},
		},
	}, &buf)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	src := buf.String()
	if !strings.Contains(src, "pub const spec = protocol.Spec{") {
		t.Fatalf("generated source is missing spec constant:\n%s", src)
	}
	if !strings.Contains(src, "pub const systems = verifier_mod.Systems{") {
		t.Fatalf("generated source is missing systems constant:\n%s", src)
	}
	if strings.Contains(src, "pub fn ") {
		t.Fatalf("generated source must not emit functions:\n%s", src)
	}
	if got := strings.Count(src, "// Round "); got != 2 {
		t.Fatalf("expected one comment per round, got %d:\n%s", got, src)
	}
	if !strings.Contains(src, ".round_coin_counts = &[_]usize{ 0, 0, 0 },") {
		t.Fatalf("generated source is missing fallback round coin counts for two rounds:\n%s", src)
	}
	if !strings.Contains(src, ".round_coin_offsets = &[_]usize{ 0, 0, 0 },") {
		t.Fatalf("generated source is missing fallback round coin offsets for two rounds:\n%s", src)
	}
	if !strings.Contains(src, `const verifier_ray = @import("verifier_ray");`) {
		t.Fatalf("generated source must use named verifier_ray import:\n%s", src)
	}
}
