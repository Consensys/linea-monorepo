package config

import (
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

const (
	// envconfig prefix
	proverPrefix        = "PROVER"
	full         string = "full"
	fullLarge    string = "full-large"
	light        string = "light"
)

// Spec for the prover configuration
type ProverSpec struct {
	PKeyFile string `envconfig:"PKEY_FILE" required:"true"`
	R1CSFile string `envconfig:"R1CS_FILE" required:"true"`
	VKeyFile string `envconfig:"VKEY_FILE" required:"true"`

	ProfilingEnabled   bool   `envconfig:"PROFILING_ENABLED"`
	TracingEnabled     bool   `envconfig:"TRACING_ENABLED"`
	SkipTraces         bool   `envconfig:"SKIP_TRACES" default:"false"`
	DevLightVersion    bool   `envconfig:"DEV_LIGHT_VERSION" default:"true"`
	Version            string `envconfig:"VERSION" required:"true"`
	ConflatedTracesDir string `envconfig:"CONFLATED_TRACES_DIR" required:"true"`
	WithStateManager   bool   `envconfig:"WITH_STATE_MANAGER" required:"false" default:"false"`
	WithKeccak         bool   `envconfig:"WITH_KECCAK" required:"false" default:"false"`
	WithEcdsa          bool   `envconfig:"WITH_ECDSA" required:"false" default:"false"`

	VerIDLight     int `envconfig:"VERIFIER_INDEX_LIGHT" required:"false" default:"0"`
	VerIDFull      int `envconfig:"VERIFIER_INDEX_FULL" required:"false" default:"1"`
	VerIDFullLarge int `envconfig:"VERIFIER_INDEX_FULL_LARGE" required:"false" default:"2"`
}

// Returns the ethereum configuration
func GetProver() (*ProverSpec, error) {
	conf := &ProverSpec{}
	err := envconfig.Process(proverPrefix, conf)
	if err != nil {
		return nil, err
	}

	if conf.Mode() == full || conf.Mode() == fullLarge {
		// For sanity, checks that the PKey, R1cs etc.. are
		// pointed to "full" and not "light".
		scsOk := strings.Contains(conf.R1CSFile, full)
		pkOk := strings.Contains(conf.PKeyFile, full)
		vkOk := strings.Contains(conf.VKeyFile, full)
		if !scsOk || !pkOk || !vkOk {
			logrus.Errorf("configuraton indicated that "+
				"the mode is %v, but the filepaths containing the "+
				"assets do not all contain the keyword %q (%v, %v, %v)",
				conf.Mode(), full, conf.R1CSFile, conf.PKeyFile, conf.VKeyFile)
		}
	}

	return conf, nil
}

// Returns the ethereum config or panic on error
func MustGetProver() *ProverSpec {
	conf, err := GetProver()
	if err != nil {
		logrus.Panicf("could not return prover config : %v", err)
	}
	return conf
}

func (p *ProverSpec) Mode() string {
	if p.DevLightVersion {
		return "light"
	}

	// Return full-large mode
	if IS_LARGE {
		return "full-large"
	}

	return "full"
}

func (p *ProverSpec) VerifierIndex() int {

	if p.DevLightVersion || p.SkipTraces {
		return p.VerIDLight
	}

	// Return full-large mode
	if IS_LARGE {
		return p.VerIDFullLarge
	}

	return p.VerIDFull
}
