package profiling

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/pprof/profile"
	"github.com/pkg/errors"
)

// GenerateFlameGraph generates a flame graph from a profile file
func (pl *PerformanceLog) GenerateFlameGraph() error {
	// Ensure the output directory exists
	if err := os.MkdirAll(pl.FlameGraphPath, 0755); err != nil {
		return errors.Wrap(err, "failed to create output directory")
	}

	// Read the profile file
	prof, err := os.Open(pl.ProfilePath)
	if err != nil {
		return errors.Wrap(err, "failed to open profile file")
	}
	defer prof.Close()

	// Parse the profile
	p, err := profile.Parse(prof)
	if err != nil {
		return errors.Wrap(err, "failed to parse profile")
	}

	// Write the profile to a temporary file
	tempFile := filepath.Join(pl.FlameGraphPath, "temp.pb.gz")
	if err := writeProfile(p, tempFile); err != nil {
		return errors.Wrap(err, "failed to write temporary profile")
	}

	// Generate the flame graph using pprof
	flameGraphPath := filepath.Join(pl.FlameGraphPath, "flamegraph.svg")
	cmd := exec.Command("go", "tool", "pprof", "-svg", "-output", flameGraphPath, tempFile)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to generate flame graph")
	}

	// Clean up the temporary file
	os.Remove(tempFile)
	return nil
}

// writeProfile writes a profile to a file
func writeProfile(p *profile.Profile, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	return p.Write(out)
}
