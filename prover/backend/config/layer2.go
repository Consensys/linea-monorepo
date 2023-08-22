package config

import (
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

const (
	// envconfig prefix used by the app
	ethPrefix = "LAYER2"
)

// Spec for the configuration of the ethereum related fields
type Layer2Spec struct {
	ChainId            int    `envconfig:"CHAIN_ID" required:"true"`
	ChainIdOverride    bool   `envconfig:"CHAIN_ID_OVERRIDE" required:"false" default:"false"`
	MsgServiceContract string `envconfig:"MESSAGE_SERVICE_CONTRACT" required:"true"`
}

// Return the configuration file
func GetLayer2() (*Layer2Spec, error) {
	conf := &Layer2Spec{}
	err := envconfig.Process(ethPrefix, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// Get the Ethereum config or panic
func MustGetLayer2() *Layer2Spec {
	eth, err := GetLayer2()
	if err != nil {
		logrus.Panicf("could not return config : %v", err)
	}
	return eth
}

// Returns the bridge address of the contract. Panic if given an
// incorrect address.
func (l *Layer2Spec) MustGetMsgServiceContract() common.Address {
	return parseAddressOrPanic(l.MsgServiceContract)
}

// Ensures the string is a correct hex address and return the address
func parseAddressOrPanic(s string) common.Address {
	b := hexutil.MustDecode(s)
	if len(b) != 20 {
		utils.Panic("the string is not a correct address, it should have exactly 20 bytes found %v", len(b))
	}
	res := common.Address{}
	copy(res[:], b)
	return res
}
