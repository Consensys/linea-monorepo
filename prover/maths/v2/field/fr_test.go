package field

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrAdd(t *testing.T) {
	a := NewFr(1)
	b := NewFr(2)
	c := NewFr(3)

	a.Add(&a, &b)

	require.Equal(t, c, a)
}

func TestFrMul(t *testing.T) {
	a := NewFr(2)
	b := NewFr(3)
	c := NewFr(6)

	a.Mul(&a, &b)

	require.Equal(t, c, a)
}

func TestFrInverse(t *testing.T) {
	a := NewFr(2)
	a.Inverse(&a)
	a.Inverse(&a)

	require.Equal(t, NewFr(2), a)
}

func TestFrExp(t *testing.T) {
	a := NewFr(2)
	b := NewFr(4)

	b.ExpToInt(a, 2)

	require.Equal(t, NewFr(4), b)
}

func TestFrRandom(t *testing.T) {
	rng := rand.New(rand.NewPCG(0, 0))
	a := PseudoRand(rng)

	require.NotNil(t, a)
}

func TestFrSetBytes(t *testing.T) {
	a := NewFr(1)
	b := NewFr(0)

	bytesA := a.Bytes()
	b.SetBytesCanonical(bytesA[:])

	require.Equal(t, a, b)
}

func TestFrSetInt(t *testing.T) {
	a := NewFr(1)
	b := NewFr(0)

	b.SetInt64(1)

	require.Equal(t, a, b)
}

func TestFrSetString(t *testing.T) {
	a := NewFr(1)
	b := NewFr(0)

	b.SetString("1")

	require.Equal(t, a, b)
}

func TestFrIsEqual(t *testing.T) {
	a := NewFr(1)
	b := NewFr(1)
	c := NewFr(2)

	require.True(t, a.Equal(&b))
	require.False(t, a.Equal(&c))
}

func TestFrIsZero(t *testing.T) {
	a := NewFr(0)
	b := NewFr(1)

	require.True(t, a.IsZero())
	require.False(t, b.IsZero())
}

func TestFrIsOne(t *testing.T) {
	a := NewFr(1)
	b := NewFr(0)

	require.True(t, a.IsOne())
	require.False(t, b.IsOne())
}

func BenchmarkFrAdd(b *testing.B) {
	a := NewFr(1)
	c := NewFr(2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Add(&a, &c)
	}
}

func BenchmarkFrMul(b *testing.B) {
	a := NewFr(2)
	c := NewFr(3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Mul(&a, &c)
	}
}

func BenchmarkFrInverse(b *testing.B) {
	a := NewFr(2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Inverse(&a)
	}
}

func BenchmarkFrExp(b *testing.B) {
	a := NewFr(2)
	c := NewFr(4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ExpToInt(a, 256)
	}
}

func BenchmarkFrRandom(b *testing.B) {
	rng := rand.New(rand.NewPCG(0, 0))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PseudoRand(rng)
	}
}

func BenchmarkFrSetBytes(b *testing.B) {
	a := NewFr(1)
	c := NewFr(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytesA := a.Bytes()
		c.SetBytesCanonical(bytesA[:])
	}
}

func BenchmarkFrSetInt(b *testing.B) {

	c := NewFr(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetInt64(1)
	}
}

func BenchmarkFrSetString(b *testing.B) {

	c := NewFr(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetString("1")
	}
}

func BenchmarkFrIsEqual(b *testing.B) {
	a := NewFr(1)
	c := NewFr(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Equal(&c)
	}
}

func BenchmarkFrIsZero(b *testing.B) {
	a := NewFr(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.IsZero()
	}
}

func BenchmarkFrIsOne(b *testing.B) {
	a := NewFr(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.IsOne()
	}
}
