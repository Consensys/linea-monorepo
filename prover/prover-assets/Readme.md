# Generating the setup

## Required infra

We run the setup on 1Tib memory machines but it should be possible to run the setup with smaller machines too. You will still need at lease 100GiB of RAM, so you will need a big machine regardless. It may change in the future.

The setup will also takes 100s of GiB of disk. So make sure you have enough space for it.

## Checkout your version

For instance,

Linea beta-v5.1 Fusaka:  `e1e35d2`

```
git checkout e1e35d2
```

## Get the KZG-SRS locally

Please contact us, if you need assistance to obtain them.

You should have them in the `prover-assets/kzgsrs`

```

~/zkevm-monorepo/prover$ tree ./prover-assets/kzgsrs

===vvvv===

./prover-assets/kzgsrs
├── kzg_srs_canonical_1027_bls12377_aleo.memdump
├── kzg_srs_canonical_1027_bn254_aztec.memdump
├── kzg_srs_canonical_1027_bw6761_celo.memdump
├── kzg_srs_canonical_1048579_bn254_aztec.memdump
├── kzg_srs_canonical_131075_bn254_aztec.memdump

[...]

├── kzg_srs_lagrange_8192_bn254_aztec.memdump
├── kzg_srs_lagrange_8192_bw6761_celo.memdump
├── kzg_srs_lagrange_8388608_bls12377_aleo.memdump
├── kzg_srs_lagrange_8388608_bn254_aztec.memdump
└── kzg_srs_lagrange_8388608_bw6761_celo.memdump

0 directories, 98 files
``` 

## Review the config

If your intend to use it for replicating the setup. You can skip this step. Otherwise, feel free to review the configs in the `config` folder. We maintain the config 
files used for each setup.

```
prover/config/config-mainnet-limitless.toml
prover/config/config-sepolia-limitless.toml
```

### Version and environment

`version` and `environment` tell the setup generator to write the setups in the `./prover-assets/<version>/<environment>` directory. We recomment you don't change the `assets_dir` because this is where the setup generator will look for the above-mentionned KZG-SRS. If you have a custom chain (without forking), it can make sense to use your own environment and stick to the same version number as us.

```
environment = "mainnet"
version = "6.1.3"
assets_dir = "./prover-assets"
```

### Layer 2 information

If you launch your own L2, the setup will need to provided the `chain_id` of your L2 and `message_service_contract` (the L2 bridge contract)

```
[layer2]
chain_id = 59144
message_service_contract = "0x508Ca82Df566dCD1B0DE8296e70a96332cD644ec"
```

### Trace limits

The trace limits are relevant for the full-prover and full-large-prover (but not for the limitless prover). In each release, we set these parameters after having done statistical analysis so that you don't have to. If you still want to experiment with it, we advise you to ensure that the "large" value of a limit is at least twice the "regular" limit.

```
[traces_limits]
ADD = 262144
BIN = 262144
BLAKE_MODEXP_DATA = 16384
[...]

[traces_limits_large]
ADD = 524288
BIN = 524288
BLAKE_MODEXP_DATA = 32768
[...]
```

### Aggregation

```
[aggregation]

# num_proofs will generate an aggregation proof circuit for each value of 
# num_proof. This helps reducing the costs if we seek to aggregate fewer 
# conflations.
num_proofs = [10, 20, 50, 100, 200, 400]

# allowed_inputs controls which circuits are white-listed in the aggregation
# proof. For production (mainnet), you must ensure that the dev circuits are
# not included in that list.
allowed_inputs = ["execution", "execution-large", "execution-limitless", "blob-decompression-v0", "blob-decompression-v1"]
```

### Public input interconnection

```
[public_input_interconnection]
# max_nb_decompression controls the number of compression range that can be 
# aggregated at once. Better to leave it as max(aggregation.num_proofs)
max_nb_decompression = 400

# max_nb_execution controls the number of compression range that can be 
# aggregated at once. Better to leave it as max(aggregation.num_proofs)
max_nb_execution = 400

# max_nb_execution controls the number of compression range that can be 
# aggregated at once. Must be at least as big as max(aggregation.num_proofs)
max_nb_circuits = 400

# The maximum number of L2->L1 messages that can be proved by an execution 
# proof at once.
execution_max_nb_msg = 16

# These control the total number of L2->L1 messages that can be proved by an
# aggregation proof at once. The value is 200 * (2**5).
l2_msg_merkle_depth = 5
l2_msg_max_nb_merkle = 200
```

## Running the setup

For replicating Linea's setup, you can run. This should take one of two hours
spits a lot of logs.

```
# to use config/config-mainnet-limitless.toml
make setup-sepolia

# to use config/config-sepolia-limitless.toml
make setup-mainnet
```

## Reviewing the generated assets

You may review the generated assets, by checking out, the verifier solidity contract.

```
# assumming you did not change the config files
less ./prover-assets/6.1.3/mainnet/emulation/Verifier.sol
```

It will differ from the contract, we actually deploy because, we apply on top of it:

* linting rules
* renaming of the contract
* change in the header solidity pragma

But otherwise, you have the same as us. In the past, we have had non-determinism 
issues with the setup: e.g. the generator may randomly order constraints in some
places but without breaking the proving flow. It usually get resolved fast, but 
if you fail regenerating the same setup it may be a good idea to retry to see
if the verification key change. This can be tracked by looking into the 
`Manifest.json` file of each asset.


