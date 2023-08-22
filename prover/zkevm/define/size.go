//go:build !largesize

package define

// Numbers of rows for the different modules
const (
	SIZE                       = 1 << 22
	SIZE_add                   = 524288
	SIZE_bin                   = 262144
	SIZE_binRT                 = 262144
	SIZE_ext                   = 16384
	SIZE_ec_data               = 4096
	SIZE_hub                   = 2097152
	SIZE_instructionsubdecoder = 512 // Ugly hack, TODO: @franklin
	SIZE_mod                   = 131072
	SIZE_mul                   = 65536
	SIZE_rom                   = 1 << 22
	SIZE_shf                   = 65536
	SIZE_shfRT                 = 262144
	SIZE_txcd_data             = 1048576 // legacy
	SIZE_mxp                   = 524288
	SIZE_hash_data             = 32768
	SIZE_log_data              = 16384
	SIZE_rlp                   = 512
	SIZE_txRlp                 = 65536 // TODO: Phoney value
	SIZE_wcp                   = 262144
)
