package main

import (
	"flag"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/ethereum/go-ethereum/common"
)

var flagStart = flag.Int("start", 0, "starting block")
var flagEnd = flag.Int("end", 0, "ending block (inclusive)")

func main() {
	flag.Parse()
	cfg := config.Config{
		Environment:                "",
		Version:                    "",
		LogLevel:                   0,
		AssetsDir:                  "",
		Controller:                 config.Controller{},
		Execution:                  config.Execution{},
		BlobDecompression:          config.BlobDecompression{},
		Aggregation:                config.Aggregation{},
		PublicInputInterconnection: config.PublicInput{},
		Debug: struct {
			Profiling bool `mapstructure:"profiling"`
			Tracing   bool `mapstructure:"tracing"`
		}{},
		Layer2: struct {
			ChainID           uint           `mapstructure:"chain_id" validate:"required"`
			MsgSvcContractStr string         `mapstructure:"message_service_contract" validate:"required,eth_addr"`
			MsgSvcContract    common.Address `mapstructure:"-"`
		}{},
		TracesLimits:      config.TracesLimits{},
		TracesLimitsLarge: config.TracesLimits{},
	}
}
