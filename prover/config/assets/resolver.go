package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/config"
)

// AssetResolver maps job names to asset file paths in the asset directory.
// Only for the execution-limitless circuit
type AssetResolver struct {
	SetupDir string
}

// NewResolver builds a resolver from your existing config.
func NewResolver(cfg *config.Config) *AssetResolver {
	return &AssetResolver{
		SetupDir: cfg.PathForSetup(string(circuits.ExecutionLimitlessCircuitID)),
	}
}

// AssetsForJob returns (allPaths, criticalPaths, error).
// criticalPaths are the minimal set that must exist (missing => error).
func (r *AssetResolver) AssetsForJob(jobName string) (allPaths []string, criticalPaths []string, err error) {
	switch jobName {
	case "bootstrap":
		return r.bootstrapAssets()
	case "conglomeration":
		return r.conglomerationAssets()
	default:
		// GL and LPP workers
		if strings.HasPrefix(jobName, "gl-") {
			module := strings.TrimPrefix(jobName, "gl-")
			return r.glAssets(module)
		}
		if strings.HasPrefix(jobName, "lpp-") {
			module := strings.TrimPrefix(jobName, "lpp-")
			return r.lppAssets(module)
		}
		// default: no assets to prefetch
		return nil, nil, fmt.Errorf("unknown limitless job name: %s", jobName)
	}
}

func (r *AssetResolver) bootstrapAssets() ([]string, []string, error) {
	required := []string{
		filepath.Join(r.SetupDir, "dw-bootstrapper.bin"),
		filepath.Join(r.SetupDir, "zkevm-wiop.bin"),
		filepath.Join(r.SetupDir, "disc.bin"),
		filepath.Join(r.SetupDir, "verification-key-merkle-tree.bin"),
	}
	glBlueprints, _ := filepath.Glob(filepath.Join(r.SetupDir, "dw-blueprint-gl-*.bin"))
	lppBlueprints, _ := filepath.Glob(filepath.Join(r.SetupDir, "dw-blueprint-lpp-*.bin"))

	all := append([]string{}, required...)
	all = append(all, glBlueprints...)
	all = append(all, lppBlueprints...)

	// check critical exist
	var missing []string
	for _, p := range required {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			missing = append(missing, p)
		}
	}
	if len(missing) > 0 {
		return all, required, fmt.Errorf("missing critical bootstrap assets: %v", missing)
	}
	return all, required, nil
}

func (r *AssetResolver) conglomerationAssets() ([]string, []string, error) {
	all := []string{
		filepath.Join(r.SetupDir, "dw-compiled-conglomeration.bin"),
		filepath.Join(r.SetupDir, "verification-key-merkle-tree.bin"),
		filepath.Join(r.SetupDir, "circuit.bin"),
		filepath.Join(r.SetupDir, "verifying_key.bin"),
		filepath.Join(r.SetupDir, "manifest.json"),
	}

	// compiled conglomeration critical
	critical := []string{all[0]}
	if _, err := os.Stat(critical[0]); os.IsNotExist(err) {
		return all, critical, fmt.Errorf("missing critical conglomeration asset: %s", critical[0])
	}
	return all, critical, nil
}

func (r *AssetResolver) glAssets(module string) ([]string, []string, error) {
	all := []string{}
	compiled := filepath.Join(r.SetupDir, "dw-compiled-gl-"+module+".bin")
	if _, err := os.Stat(compiled); err == nil {
		all = append(all, compiled)
	}

	// if compiled present, consider it critical (caller can decide behaviour)
	return all, all, nil
}

func (r *AssetResolver) lppAssets(module string) ([]string, []string, error) {
	all := []string{}
	compiled := filepath.Join(r.SetupDir, "dw-compiled-lpp-"+module+".bin")
	if _, err := os.Stat(compiled); err == nil {
		all = append(all, compiled)
	}

	return all, all, nil
}
