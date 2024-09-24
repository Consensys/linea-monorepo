package serialization

import (
	"reflect"

	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

var _registryBkp = collection.NewMapping[string, reflect.Type]()

func backupRegistryAndReset() {
	_registryBkp = implementationRegistry
	implementationRegistry = collection.NewMapping[string, reflect.Type]()
}

func restoreRegistryFromBackup() {
	implementationRegistry = _registryBkp
}
