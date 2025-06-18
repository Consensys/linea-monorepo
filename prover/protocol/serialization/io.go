package serialization

import (
	"bytes"
	"fmt"
	"path"

	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// Serializes the object and write the serialized data to a fileName specified in filePath
func SerializeAndWrite(filePath string, fileName string, object any, reader *bytes.Reader) error {
	data, err := Serialize(object)
	if err != nil {
		return fmt.Errorf("failed to serialize %s: %w", fileName, err)
	}
	reader.Reset(data)
	fullPath := path.Join(filePath, fileName)
	if err := utils.WriteToFile(fullPath, reader); err != nil {
		return fmt.Errorf("failed to write %s: %w", fullPath, err)
	}
	logrus.Infof("Written %s to %s", fileName, fullPath)
	return nil
}

// Reads the serialized data from the file and deserialize it into the target object
func ReadAndDeserialize(filePath string, fileName string, target any, readBuf *bytes.Buffer) error {
	readBuf.Reset()
	fullPath := path.Join(filePath, fileName)
	if err := utils.ReadFromFile(fullPath, readBuf); err != nil {
		return fmt.Errorf("failed to read %s: %w", fullPath, err)
	}
	if err := Deserialize(readBuf.Bytes(), target); err != nil {
		return fmt.Errorf("failed to deserialize %s: %w", fileName, err)
	}
	logrus.Infof("Read and deserialized %s from %s", fileName, fullPath)
	return nil
}
