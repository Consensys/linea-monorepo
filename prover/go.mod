module github.com/consensys/linea-monorepo/prover

go 1.24.0

toolchain go1.24.6

require (
	github.com/bits-and-blooms/bitset v1.20.0
	github.com/consensys/bavard v0.1.31-0.20250314194434-b30d4344e6d4
	github.com/consensys/compress v0.2.5
	github.com/consensys/gnark v0.12.1-0.20250501002417-facdd9882b80
	github.com/consensys/gnark-crypto v0.17.1-0.20250326164229-5fd6610ac2a1
	github.com/consensys/go-corset v1.0.12-0.20250729080012-3d83adbcfe23
	github.com/crate-crypto/go-kzg-4844 v1.1.0
	github.com/dlclark/regexp2 v1.11.2
	github.com/dnlo/struct2csv v0.0.0-20190928115744-2f584471b24e
	github.com/fxamacker/cbor/v2 v2.7.0
	github.com/go-playground/assert/v2 v2.2.0
	github.com/go-playground/validator/v10 v10.22.0
	github.com/gofrs/flock v0.12.0
	github.com/google/uuid v1.6.0
	github.com/icza/bitio v1.1.0
	github.com/leanovate/gopter v0.2.11
	github.com/pierrec/lz4/v4 v4.1.22
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.19.1
	github.com/rs/zerolog v1.33.0
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.10.0
	golang.org/x/crypto v0.39.0
	golang.org/x/net v0.38.0
	golang.org/x/sync v0.15.0
	golang.org/x/time v0.5.0
)

require (
	github.com/VictoriaMetrics/fastcache v1.12.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/crate-crypto/go-ipa v0.0.0-20240223125850-b1e8a79f509c // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/ethereum/c-kzg-4844 v1.0.3 // indirect
	github.com/ethereum/go-verkle v0.1.1-0.20240829091221-dffa7562dbe9 // indirect
	github.com/felixge/fgprof v0.9.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.4 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/pprof v0.0.0-20240727154555-813a5fbdbec8 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/holiman/uint256 v1.3.1 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/ingonyama-zk/icicle-gnark/v3 v3.2.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/ronanh/intcomp v1.1.0 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/supranational/blst v0.3.13 // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.8.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/ethereum/go-ethereum v1.14.13
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/pkg/profile v1.7.0
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/exp v0.0.0-20240823005443-9b4947da3948
	golang.org/x/sys v0.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
