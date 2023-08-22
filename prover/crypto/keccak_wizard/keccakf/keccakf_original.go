// this is original keccakf (it just computes the output)
package keccak

// rc stores the round constants for use in the ι step.
var rc = [24]uint64{
	0x0000000000000001,
	0x0000000000008082,
	0x800000000000808A,
	0x8000000080008000,
	0x000000000000808B,
	0x0000000080000001,
	0x8000000080008081,
	0x8000000000008009,
	0x000000000000008A,
	0x0000000000000088,
	0x0000000080008009,
	0x000000008000000A,
	0x000000008000808B,
	0x800000000000008B,
	0x8000000000008089,
	0x8000000000008003,
	0x8000000000008002,
	0x8000000000000080,
	0x000000000000800A,
	0x800000008000000A,
	0x8000000080008081,
	0x8000000000008080,
	0x0000000080000001,
	0x8000000080008008,
}
var lr = [5][5]uint64{
	{0, 36, 3, 41, 18},
	{1, 44, 10, 45, 2},
	{62, 6, 43, 15, 61},
	{28, 55, 25, 21, 56},
	{27, 20, 39, 8, 14},
}

// keccakF1600 applies the Keccak permutation to a 1600b-state
func KeccakF1600Original(a [5][5]uint64) (output [5][5]uint64) {
	var aTheta, aPi [5][5]uint64
	var c, d [5]uint64

	for l := 0; l < 24; l++ {
		for i := 0; i < 5; i++ {
			c[i] = a[i][0] ^ a[i][1] ^ a[i][2] ^ a[i][3] ^ a[i][4]

		}

		for i := 0; i < 5; i++ {
			d[i] = c[(i-1+5)%5] ^ (c[(i+1)%5]<<1 | c[(i+1)%5]>>63)
		}
		//Theta step
		for i := 0; i < 5; i++ {
			for j := 0; j < 5; j++ {
				aTheta[i][j] = a[i][j] ^ d[i]

			}
		}
		//Pi and Rho step
		for i := 0; i < 5; i++ {
			for j := 0; j < 5; j++ {
				aPi[j][(2*i+3*j)%5] = aTheta[i][j]<<lr[i][j] | aTheta[i][j]>>(64-lr[i][j])

			}
		}
		//Chi nad Iota step
		for i := 0; i < 5; i++ {
			for j := 0; j < 5; j++ {
				if i == 0 && j == 0 {
					a[i][j] = aPi[i][j] ^ (aPi[(i+2)%5][j] &^ aPi[(i+1)%5][j]) ^ rc[l]
				} else {
					a[i][j] = aPi[i][j] ^ (aPi[(i+2)%5][j] &^ aPi[(i+1)%5][j])
				}
			}
		}
	}
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			output[i][j] = a[i][j]
		}
	}

	return output
}
