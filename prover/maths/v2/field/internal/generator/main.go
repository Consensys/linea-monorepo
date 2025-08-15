package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {

	configFileFlag := flag.String("config", "", "path to the config file")
	flag.Parse()

	if *configFileFlag == "" {
		panic("missing config file in the command arguments")
	}

	cfgFile, err := os.ReadFile(*configFileFlag)
	if err != nil {
		panic(fmt.Sprintf("could not read the config file: %v", err))
	}

	cfg := &GeneratorConfig{}
	if err := json.Unmarshal(cfgFile, cfg); err != nil {
		panic(fmt.Sprintf("could not unmarshal the config file: %v", err))
	}

	if len(cfg.GeneratedFilename) == 0 {
		panic("missing generated filename in the config file")
	}

	if len(cfg.Mapping) == 0 {
		panic("missing mappings in the config file")
	}

	if len(cfg.NewPackageName) == 0 {
		panic("missing new package name in the config file")
	}

	generatedFile, err := os.Create(cfg.GeneratedFilename)
	if err != nil {
		panic(fmt.Sprintf("could not open the generated file: %v", err))
	}

	defer generatedFile.Close()

	if err := Generate(generatedFile, cfg); err != nil {
		panic(err)
	}
}
