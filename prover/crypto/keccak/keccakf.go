package keccak

// Dimension of the keccak state matrix. The State is structured
// as an array of 5 x 5 uint64.
const Dim = 5           // size of the keccak state matrix
const NumRound int = 24 // number of rounds in the keccak permutation

var LR = [Dim][Dim]int{
	{0, 36, 3, 41, 18},
	{1, 44, 10, 45, 2},
	{62, 6, 43, 15, 61},
	{28, 55, 25, 21, 56},
	{27, 20, 39, 8, 14},
}

// RC stores the rounds constants for the \iota steps
var RC = [NumRound]uint64{
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

// Permute applies the Keccakf permutation over the state and appends an entry
// to the PermTraces function.
func (a *State) Permute(traces *PermTraces) {

	// Save the input of the permutation
	if traces != nil {
		traces.KeccakFInps = append(traces.KeccakFInps, *a)
	}

	for round := 0; round < NumRound; round++ {
		a.ApplyKeccakfRound(round)
	}

	// Save the output of the permutation
	if traces != nil {
		traces.KeccakFOuts = append(traces.KeccakFOuts, *a)
	}
}

// ApplyKeccakfRound applies one round of the keccakf permutation
func (a *State) ApplyKeccakfRound(round int) {
	a.Theta()
	a.Rho()
	b := a.Pi()
	a.Chi(&b)
	a[0] ^= RC[round]
}

// Applies the theta step over the state
// See https://keccak.team/keccak_specs_summary.html
//
//	C[x] = A[x,0] xor A[x,1] xor A[x,2] xor A[x,3] xor A[x,4],   for x in 0…4
//	D[x] = C[x-1] xor rot(C[x+1],1),                             for x in 0…4
//	A[x,y] = A[x,y] xor D[x],                           for (x,y) in (0…4,0…4)
func (a *State) Theta() {
	var c0, c1, c2, c3, c4 uint64
	var d0, d1, d2, d3, d4 uint64

	c0 = a[0] ^ a[5] ^ a[10] ^ a[15] ^ a[20]
	c1 = a[1] ^ a[6] ^ a[11] ^ a[16] ^ a[21]
	c2 = a[2] ^ a[7] ^ a[12] ^ a[17] ^ a[22]
	c3 = a[3] ^ a[8] ^ a[13] ^ a[18] ^ a[23]
	c4 = a[4] ^ a[9] ^ a[14] ^ a[19] ^ a[24]

	d0 = c4 ^ (c1<<1 | c1>>63)
	d1 = c0 ^ (c2<<1 | c2>>63)
	d2 = c1 ^ (c3<<1 | c3>>63)
	d3 = c2 ^ (c4<<1 | c4>>63)
	d4 = c3 ^ (c0<<1 | c0>>63)

	a[0] ^= d0
	a[5] ^= d0
	a[10] ^= d0
	a[15] ^= d0
	a[20] ^= d0

	a[1] ^= d1
	a[6] ^= d1
	a[11] ^= d1
	a[16] ^= d1
	a[21] ^= d1

	a[2] ^= d2
	a[7] ^= d2
	a[12] ^= d2
	a[17] ^= d2
	a[22] ^= d2

	a[3] ^= d3
	a[8] ^= d3
	a[13] ^= d3
	a[18] ^= d3
	a[23] ^= d3

	a[4] ^= d4
	a[9] ^= d4
	a[14] ^= d4
	a[19] ^= d4
	a[24] ^= d4
}

// Applies the \rho and \pi steps
//
// B[y,2*x+3*y] = rot(A[x,y], r[x,y]),  for (x,y) in (0…4,0…4)
func (a *State) Rho() {

	a[5] = a[5]<<36 | a[5]>>28
	a[10] = a[10]<<3 | a[10]>>61
	a[15] = a[15]<<41 | a[15]>>23
	a[20] = a[20]<<18 | a[20]>>46

	a[1] = a[1]<<1 | a[1]>>63
	a[6] = a[6]<<44 | a[6]>>20
	a[11] = a[11]<<10 | a[11]>>54
	a[16] = a[16]<<45 | a[16]>>19
	a[21] = a[21]<<2 | a[21]>>62

	a[2] = a[2]<<62 | a[2]>>2
	a[7] = a[7]<<6 | a[7]>>58
	a[12] = a[12]<<43 | a[12]>>21
	a[17] = a[17]<<15 | a[17]>>49
	a[22] = a[22]<<61 | a[22]>>3

	a[3] = a[3]<<28 | a[3]>>36
	a[8] = a[8]<<55 | a[8]>>9
	a[13] = a[13]<<25 | a[13]>>39
	a[18] = a[18]<<21 | a[18]>>43
	a[23] = a[23]<<56 | a[23]>>8

	a[4] = a[4]<<27 | a[4]>>37
	a[9] = a[9]<<20 | a[9]>>44
	a[14] = a[14]<<39 | a[14]>>25
	a[19] = a[19]<<8 | a[19]>>56
	a[24] = a[24]<<14 | a[24]>>50
}

// Applies the \pi steps
func (a *State) Pi() (b State) {
	b[0] = a[0]
	b[16] = a[5]
	b[7] = a[10]
	b[23] = a[15]
	b[14] = a[20]
	b[10] = a[1]
	b[1] = a[6]
	b[17] = a[11]
	b[8] = a[16]
	b[24] = a[21]
	b[20] = a[2]
	b[11] = a[7]
	b[2] = a[12]
	b[18] = a[17]
	b[9] = a[22]
	b[5] = a[3]
	b[21] = a[8]
	b[12] = a[13]
	b[3] = a[18]
	b[19] = a[23]
	b[15] = a[4]
	b[6] = a[9]
	b[22] = a[14]
	b[13] = a[19]
	b[4] = a[24]
	return b
}

// Applies the Chi step
//
// A[x,y] = B[x,y] xor ((not B[x+1,y]) and B[x+2,y]),  for (x,y) in (0…4,0…4)
func (a *State) Chi(b *State) {
	a[0] = b[0] ^ (^b[1] & b[2])
	a[5] = b[5] ^ (^b[6] & b[7])
	a[10] = b[10] ^ (^b[11] & b[12])
	a[15] = b[15] ^ (^b[16] & b[17])
	a[20] = b[20] ^ (^b[21] & b[22])
	a[1] = b[1] ^ (^b[2] & b[3])
	a[6] = b[6] ^ (^b[7] & b[8])
	a[11] = b[11] ^ (^b[12] & b[13])
	a[16] = b[16] ^ (^b[17] & b[18])
	a[21] = b[21] ^ (^b[22] & b[23])
	a[2] = b[2] ^ (^b[3] & b[4])
	a[7] = b[7] ^ (^b[8] & b[9])
	a[12] = b[12] ^ (^b[13] & b[14])
	a[17] = b[17] ^ (^b[18] & b[19])
	a[22] = b[22] ^ (^b[23] & b[24])
	a[3] = b[3] ^ (^b[4] & b[0])
	a[8] = b[8] ^ (^b[9] & b[5])
	a[13] = b[13] ^ (^b[14] & b[10])
	a[18] = b[18] ^ (^b[19] & b[15])
	a[23] = b[23] ^ (^b[24] & b[20])
	a[4] = b[4] ^ (^b[0] & b[1])
	a[9] = b[9] ^ (^b[5] & b[6])
	a[14] = b[14] ^ (^b[10] & b[11])
	a[19] = b[19] ^ (^b[15] & b[16])
	a[24] = b[24] ^ (^b[20] & b[21])
}

func (a *State) Iota(round int) {
	a[0] ^= RC[round]
}
