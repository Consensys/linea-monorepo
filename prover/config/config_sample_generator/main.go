package main

import (
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/consensys/linea-monorepo/prover/backend/files"
)

// tmplFile is the filepath of the template to be filled by the presently
// written generator.
const tmplFile string = "./config_sample_generator/config.go.tmpl"

// Params collects the config template parameters
type Params struct {
	Environment Environment
	Mode        Mode
	Filename    string
}

// Environment stores the environment parameters of the config
type Environment struct {
	Name                   string
	ChainID                int
	BaseFee                int
	CoinBase               string
	MessageServiceContract string
}

const (
	ExecutionDummy = 1 << iota
	DataAvailabilityDummy
	EmulationDummy
	ExecutionFull
	ExecutionFullLarge
	ExecutionLimitless
	DataAvailabilityV2
)

// Mode stores the operation modes of the prover
type Mode struct {
	ExecutionMode        string
	DataAvailabilityMode string
	AggregationMode      string
	AllowedCircuits      int
	AggregationNumProofs []int
}

var (
	// The environment parameters of the E2E tests
	e2eEnvironment = Environment{
		Name:                   "e2e",
		ChainID:                1337,
		BaseFee:                7,
		CoinBase:               "0x8F81e2E3F8b46467523463835F965fFE476E1c9E",
		MessageServiceContract: "0xe537D669CA013d86EBeF1D64e40fC74CADC91987",
	}

	// The environment parameters for devnet
	devnetEnvironment = Environment{
		Name:                   "devnet",
		ChainID:                59139,
		BaseFee:                7,
		CoinBase:               "0x19bf28626BE6f6aE4ca7d41A5aDe0305e9DC5FCA",
		MessageServiceContract: "0x33bf916373159a8c1b54b025202517bfdbb7863d",
	}

	// The environment parameters for sepolia
	sepoliaEnvironment = Environment{
		Name:                   "sepolia",
		ChainID:                59141,
		BaseFee:                7,
		CoinBase:               "0xA27342f1b74c0cfB2cda74bac1628d0C1A9752f2",
		MessageServiceContract: "0x971e727e956690b9957be6d51Ec16E73AcAC83A7",
	}

	// The environment parameters for mainnet
	mainnetEnvironment = Environment{
		Name:                   "mainnet",
		ChainID:                59144,
		BaseFee:                7,
		CoinBase:               "0x8F81e2E3F8b46467523463835F965fFE476E1c9E",
		MessageServiceContract: "0x508Ca82Df566dCD1B0DE8296e70a96332cD644ec",
	}

	// The parameters for productions setting
	productionMode = Mode{
		ExecutionMode:        "execution",
		DataAvailabilityMode: "data-availability",
		AggregationMode:      "aggregation",
		AllowedCircuits:      ExecutionFull | ExecutionFullLarge | ExecutionLimitless | DataAvailabilityV2,
		AggregationNumProofs: []int{10, 20, 50, 100, 200, 400},
	}

	// The parameters for heavier testing (for the setup of devnet and sepolia)
	stagingMode = Mode{
		ExecutionMode:        "partial",
		DataAvailabilityMode: "full",
		AggregationMode:      "full",
		AllowedCircuits:      ExecutionDummy | DataAvailabilityDummy | ExecutionFull | ExecutionFullLarge | ExecutionLimitless | DataAvailabilityV2,
		AggregationNumProofs: []int{10, 20, 50, 100, 200, 400},
	}

	// The parameters for quick light-testing setting
	devMode = Mode{
		ExecutionMode:        "dev",
		DataAvailabilityMode: "dev",
		AggregationMode:      "dev",
		AllowedCircuits:      ExecutionDummy | DataAvailabilityDummy,
		AggregationNumProofs: []int{10},
	}
)

var params = []Params{
	{
		Filename:    "config-e2e-dev.toml",
		Mode:        devMode,
		Environment: e2eEnvironment,
	},
	{
		Filename:    "config-e2e-staging.toml",
		Mode:        stagingMode,
		Environment: e2eEnvironment,
	},
	{
		Filename:    "config-devnet-dev.toml",
		Mode:        devMode,
		Environment: devnetEnvironment,
	},
	{
		Filename:    "config-devnet-staging.toml",
		Mode:        stagingMode,
		Environment: devnetEnvironment,
	},
	{
		Filename:    "config-sepolia-staging.toml",
		Mode:        stagingMode,
		Environment: sepoliaEnvironment,
	},
	{
		Filename:    "config-mainnet-staging.toml",
		Mode:        stagingMode,
		Environment: mainnetEnvironment,
	},
	{
		Filename:    "config-mainnet-prod.toml",
		Mode:        productionMode,
		Environment: mainnetEnvironment,
	},
}

func main() {

	pwd, _ := os.Getwd()
	fmt.Printf("[config] generator running from `%v`", pwd)

	tmplF := files.MustRead(tmplFile)

	tmplBytes, err1 := io.ReadAll(tmplF)
	if err1 != nil {
		panic(err1)
	}

	tmpl := template.Must(template.New("config").Parse(string(tmplBytes)))
	for _, param := range params {
		w := files.MustOverwrite(param.Filename)
		if err := tmpl.Execute(w, param); err != nil {
			panic(err)
		}
	}

}
