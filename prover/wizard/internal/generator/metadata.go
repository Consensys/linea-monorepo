package main

import (
	"fmt"
	"os"

	"github.com/consensys/bavard"
)

type params struct {
	RcvrName          string
	TypeName          string
	FileName          string
	SkipID            bool // skipID skips the generation of the id() method
	IsNoResultQuery   bool
	IsQueryWithResult bool
}

var metadataList = []params{
	{
		TypeName: "ColNatural",
		RcvrName: "nat",
		FileName: "column_natural_metadata.go",
	},
	{
		TypeName: "CoinField",
		RcvrName: "c",
		FileName: "coin_field_metadata.go",
	},
	{
		TypeName: "runtimeVerifierAction",
		RcvrName: "va",
		FileName: "runtime_verifier_action_metadata.go",
		SkipID:   true,
	},
	{
		TypeName: "runtimeProverAction",
		RcvrName: "pa",
		FileName: "runtime_prover_action_metadata.go",
		SkipID:   true,
	},
	{
		TypeName:        "QueryGlobal",
		RcvrName:        "i",
		FileName:        "query_global_metadata.go",
		IsNoResultQuery: true,
	},
	{
		TypeName:          "QueryInnerProduct",
		RcvrName:          "i",
		FileName:          "query_innerproduct_metadata.go",
		IsQueryWithResult: true,
	},
	{
		TypeName:        "QueryInclusion",
		RcvrName:        "i",
		FileName:        "query_inclusion_metadata.go",
		IsNoResultQuery: true,
	},
	{
		TypeName:        "QueryPermutation",
		RcvrName:        "i",
		FileName:        "query_permutation_metadata.go",
		IsNoResultQuery: true,
	},
	{
		TypeName:        "QueryFixedPermutation",
		RcvrName:        "i",
		FileName:        "query_fixed_permutation_metadata.go",
		IsNoResultQuery: true,
	},
	{
		TypeName:        "QueryLocalConstraint",
		RcvrName:        "i",
		FileName:        "query_local_constraint_metadata.go",
		IsNoResultQuery: true,
	},
	{
		TypeName:          "QueryLocalOpening",
		RcvrName:          "l",
		FileName:          "query_local_opening_metadata.go",
		IsQueryWithResult: true,
	},
	{
		TypeName:        "QueryMiMC",
		RcvrName:        "i",
		FileName:        "query_mimc_metadata.go",
		IsNoResultQuery: true,
	},
	{
		TypeName:        "QueryRange",
		RcvrName:        "i",
		FileName:        "query_range_metadata.go",
		IsNoResultQuery: true,
	},
	{
		TypeName:          "QueryUnivariateEval",
		RcvrName:          "u",
		FileName:          "query_univariate_metadata.go",
		IsQueryWithResult: true,
	},
}

func genMetadata() {
	for _, param := range metadataList {

		var (
			src    = "./internal/generator/metadata.go.tmpl"
			target = "./" + param.FileName
		)

		err := bavard.GenerateFromFiles(
			target,
			[]string{src},
			param,
		)

		if err != nil {
			fmt.Printf("err = %v\n", err.Error())
			os.Exit(1)
		}
	}
}
