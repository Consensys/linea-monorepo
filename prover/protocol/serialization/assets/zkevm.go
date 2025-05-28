package assets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecarith"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecpair"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/sha2"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/modexp"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager"
)

// rawZkEvm represents the serialized form of zkevm.ZkEvm.
type rawZkEvm struct {
	Arithmetization json.RawMessage `json:"arithmetization"`
	Keccak          json.RawMessage `json:"keccak"`
	StateManager    json.RawMessage `json:"stateManager"`
	PublicInput     json.RawMessage `json:"publicInput"`
	Ecdsa           json.RawMessage `json:"ecdsa"`
	Modexp          json.RawMessage `json:"modexp"`
	Ecadd           json.RawMessage `json:"ecadd"`
	Ecmul           json.RawMessage `json:"ecmul"`
	Ecpair          json.RawMessage `json:"ecpair"`
	Sha2            json.RawMessage `json:"sha2"`
	WizardIOP       json.RawMessage `json:"wizardIOP"`
}

// SerializeZkEVM serializes a zkevm.ZkEvm instance field-by-field.
func SerializeZkEVM(z *zkevm.ZkEvm) ([]byte, error) {
	if z == nil {
		return []byte(serialization.NilString), nil
	}

	raw := &rawZkEvm{}

	// Serialize Arithmetization
	if z.Arithmetization != nil {
		arithSer, err := serialization.SerializeValue(reflect.ValueOf(z.Arithmetization), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Arithmetization: %w", err)
		}
		raw.Arithmetization = arithSer
	} else {
		raw.Arithmetization = []byte(serialization.NilString)
	}

	// Serialize Keccak
	if z.Keccak != nil {
		keccakSer, err := serialization.SerializeValue(reflect.ValueOf(z.Keccak), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Keccak: %w", err)
		}
		raw.Keccak = keccakSer
	} else {
		raw.Keccak = []byte(serialization.NilString)
	}

	// Serialize StateManager
	if z.StateManager != nil {
		smSer, err := serialization.SerializeValue(reflect.ValueOf(z.StateManager), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize StateManager: %w", err)
		}
		raw.StateManager = smSer
	} else {
		raw.StateManager = []byte(serialization.NilString)
	}

	// Serialize Ecdsa
	if z.Ecdsa != nil {
		ecdsaSer, err := serialization.SerializeValue(reflect.ValueOf(z.Ecdsa), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Ecdsa: %w", err)
		}
		raw.Ecdsa = ecdsaSer
	} else {
		raw.Ecdsa = []byte(serialization.NilString)
	}

	// Serialize Modexp
	if z.Modexp != nil {
		modexpSer, err := serialization.SerializeValue(reflect.ValueOf(z.Modexp), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Modexp: %w", err)
		}
		raw.Modexp = modexpSer
	} else {
		raw.Modexp = []byte(serialization.NilString)
	}

	// Serialize Ecadd
	if z.Ecadd != nil {
		ecaddSer, err := serialization.SerializeValue(reflect.ValueOf(z.Ecadd), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Ecadd: %w", err)
		}
		raw.Ecadd = ecaddSer
	} else {
		raw.Ecadd = []byte(serialization.NilString)
	}

	// Serialize Ecmul
	if z.Ecmul != nil {
		ecmulSer, err := serialization.SerializeValue(reflect.ValueOf(z.Ecmul), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Ecmul: %w", err)
		}
		raw.Ecmul = ecmulSer
	} else {
		raw.Ecmul = []byte(serialization.NilString)
	}

	// Serialize Ecpair
	if z.Ecpair != nil {
		ecpairSer, err := serialization.SerializeValue(reflect.ValueOf(z.Ecpair), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Ecpair: %w", err)
		}
		raw.Ecpair = ecpairSer
	} else {
		raw.Ecpair = []byte(serialization.NilString)
	}

	// Serialize Sha2
	if z.Sha2 != nil {
		sha2Ser, err := serialization.SerializeValue(reflect.ValueOf(z.Sha2), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Sha2: %w", err)
		}
		raw.Sha2 = sha2Ser
	} else {
		raw.Sha2 = []byte(serialization.NilString)
	}

	// Serialize WizardIOP
	if z.WizardIOP != nil {
		iopSer, err := serialization.SerializeCompiledIOP(z.WizardIOP)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize WizardIOP: %w", err)
		}
		raw.WizardIOP = iopSer
	} else {
		raw.WizardIOP = []byte(serialization.NilString)
	}

	// Serialize PublicInput
	if z.PublicInput != nil {
		piSer, err := serialization.SerializeValue(reflect.ValueOf(z.PublicInput), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize PublicInput: %w", err)
		}
		raw.PublicInput = piSer
	} else {
		raw.PublicInput = []byte(serialization.NilString)
	}

	return serialization.SerializeAnyWithCborPkg(raw)
}

// DeserializeZkEVM deserializes a zkevm.ZkEvm instance from CBOR-encoded data.
func DeserializeZkEVM(data []byte) (*zkevm.ZkEvm, error) {
	if bytes.Equal(data, []byte(serialization.NilString)) {
		return nil, nil
	}

	var raw rawZkEvm
	if err := serialization.DeserializeAnyWithCborPkg(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to deserialize raw ZkEvm: %w", err)
	}

	z := &zkevm.ZkEvm{}
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize Arithmetization
	if !bytes.Equal(raw.Arithmetization, []byte(serialization.NilString)) {
		val, err := serialization.DeserializeValue(raw.Arithmetization, serialization.DeclarationMode, reflect.TypeOf(&arithmetization.Arithmetization{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Arithmetization: %w", err)
		}
		z.Arithmetization = val.Interface().(*arithmetization.Arithmetization)
	}

	// Deserialize WizardIOP
	if !bytes.Equal(raw.WizardIOP, []byte(serialization.NilString)) {
		iop, err := serialization.DeserializeCompiledIOP(raw.WizardIOP)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize WizardIOP: %w", err)
		}
		z.WizardIOP = iop

		// Important: Set comp to the deserilizaed IOP containing all the columns and coins
		comp = iop
	}

	// Deserialize Keccak
	if !bytes.Equal(raw.Keccak, []byte(serialization.NilString)) {
		val, err := serialization.DeserializeValue(raw.Keccak, serialization.DeclarationMode, reflect.TypeOf(&keccak.KeccakZkEVM{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Keccak: %w", err)
		}
		z.Keccak = val.Interface().(*keccak.KeccakZkEVM)
	}

	// Deserialize StateManager
	if !bytes.Equal(raw.StateManager, []byte(serialization.NilString)) {
		val, err := serialization.DeserializeValue(raw.StateManager, serialization.DeclarationMode, reflect.TypeOf(&statemanager.StateManager{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize StateManager: %w", err)
		}
		z.StateManager = val.Interface().(*statemanager.StateManager)
	}

	// Deserialize Ecdsa
	if !bytes.Equal(raw.Ecdsa, []byte(serialization.NilString)) {
		val, err := serialization.DeserializeValue(raw.Ecdsa, serialization.DeclarationMode, reflect.TypeOf(&ecdsa.EcdsaZkEvm{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Ecdsa: %w", err)
		}
		z.Ecdsa = val.Interface().(*ecdsa.EcdsaZkEvm)
	}

	// Deserialize Modexp
	if !bytes.Equal(raw.Modexp, []byte(serialization.NilString)) {
		val, err := serialization.DeserializeValue(raw.Modexp, serialization.DeclarationMode, reflect.TypeOf(&modexp.Module{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Modexp: %w", err)
		}
		z.Modexp = val.Interface().(*modexp.Module)
	}

	// Deserialize Ecadd
	if !bytes.Equal(raw.Ecadd, []byte(serialization.NilString)) {
		val, err := serialization.DeserializeValue(raw.Ecadd, serialization.DeclarationMode, reflect.TypeOf(&ecarith.EcAdd{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Ecadd: %w", err)
		}
		z.Ecadd = val.Interface().(*ecarith.EcAdd)
	}

	// Deserialize Ecmul
	if !bytes.Equal(raw.Ecmul, []byte(serialization.NilString)) {
		val, err := serialization.DeserializeValue(raw.Ecmul, serialization.DeclarationMode, reflect.TypeOf(&ecarith.EcMul{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Ecmul: %w", err)
		}
		z.Ecmul = val.Interface().(*ecarith.EcMul)
	}

	// Deserialize Ecpair
	if !bytes.Equal(raw.Ecpair, []byte(serialization.NilString)) {
		val, err := serialization.DeserializeValue(raw.Ecpair, serialization.DeclarationMode, reflect.TypeOf(&ecpair.ECPair{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Ecpair: %w", err)
		}
		z.Ecpair = val.Interface().(*ecpair.ECPair)
	}

	// Deserialize Sha2
	if !bytes.Equal(raw.Sha2, []byte(serialization.NilString)) {
		val, err := serialization.DeserializeValue(raw.Sha2, serialization.DeclarationMode, reflect.TypeOf(&sha2.Sha2SingleProvider{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Sha2: %w", err)
		}
		z.Sha2 = val.Interface().(*sha2.Sha2SingleProvider)
	}

	// Deserialize PublicInput
	if !bytes.Equal(raw.PublicInput, []byte(serialization.NilString)) {
		val, err := serialization.DeserializeValue(raw.PublicInput, serialization.DeclarationMode, reflect.TypeOf(&publicInput.PublicInput{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize PublicInput: %w", err)
		}
		z.PublicInput = val.Interface().(*publicInput.PublicInput)
	}
	return z, nil
}
