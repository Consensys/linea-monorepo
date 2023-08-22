//go:build largesize

package define

// Numbers of rows for the different modules
const (
	SIZE                       = 1 << 23
	SIZE_add                   = 1 << 20
	SIZE_bin                   = 1 << 19
	SIZE_binRT                 = 1 << 19
	SIZE_ext                   = 1 << 15
	SIZE_ec_data               = 1 << 12
	SIZE_hash_data             = 1 << 15
	SIZE_hub                   = 1 << 22
	SIZE_log_data              = 1 << 14
	SIZE_mod                   = 1 << 17
	SIZE_mul                   = 1 << 17
	SIZE_mxp                   = 1 << 20
	SIZE_rom                   = 1 << 23
	SIZE_shf                   = 1 << 17
	SIZE_shfRT                 = 1 << 17
	SIZE_txcd_data             = 1048576 // legacy
	SIZE_wcp                   = 1 << 19
	SIZE_instructionsubdecoder = 1 << 9
	SIZE_rlp                   = 1 << 10
	SIZE_txRlp                 = 1 << 19 // TODO: Phoney value
)
