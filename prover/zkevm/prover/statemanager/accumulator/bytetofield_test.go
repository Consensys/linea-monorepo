package accumulator

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestByteToField(t *testing.T) {
	b1 := []byte{0x00, 0x01, 0x00, 0x03}
	b2 := []byte{0x00, 0x01}
	b3 := []byte{0x00, 0x03}
	f1, _ := byteToField(b1)
	f2, _ := byteToField(b2)
	f3, _ := byteToField(b3)
	logrus.Printf("f1 = %v", f1.String())
	logrus.Printf("f2 = %v", f2.String())
	logrus.Printf("f3 = %v", f3.String())
	multOffset := 1 << 16
	f4 := field.NewElement(uint64(multOffset))
	f2.Mul(&f2, &f4)
	assert.Equal(t, &f1, f3.Add(&f3, &f2))
}
