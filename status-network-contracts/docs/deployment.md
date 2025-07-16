## Deployment steps

1. Deploy StakeManager

```
MNEMONIC=$YOUR_MNEMONIC forge script scripts/DeployStakeManager.s.sol --rpc-url $RPC_URL --broadcast
```

2. Verify StakeManagerProxy

2.1 First, we need to find the **implementation** address that the proxy points to:

```sh
$ cast storage $STAKE_MANAGER_PROXY_ADDRESS 0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc --rpc-url $RPC_SN_SEPOLIA
```

2.2 Then, ABI encode the constructor arguments for `TransparentProxy`:

```sh
$ cast calldata "initialize(address,address)" $OWNER_ADDRESS $SNT_ADDRESS
```

Where `$OWNER` is the deployer. This will return the initialization calldata to verify `StakeManagerProxy`:

2.3 Then, verify the contract:

```sh
$ forge verify-contract \
--chain-id 1660990954 \
--num-of-optimizations 10000 \
--watch \
--compiler-version $COMPILER_VERSION \
$PROXY_CONTRACT_ADDRESS \
--rpc-url $RPC_SN_SEPOLIA \
--verifier-url $VERIFIER_URL_SN_SEPOLIA
--verifier blockscout \
--constructor-args $(cast abi-encode "constructor(address,bytes)" $STAKE_MANAGER_IMPL_ADDRESS $INITIALIZE_CALLDATA)
```

3. Verify `StakeManager`

```sh
$ forge verify-contract \
--chain-id 1660990954 \
--num-of-optimizations 10000 \
--watch \
--compiler-version $COMPILER_VERSION \
$STAKE_MANAGER_CONTRACT_ADDRESS \
--rpc-url $RPC_SN_SEPOLIA \
--verifier-url $VERIFIER_URL_SN_SEPOLIA
--verifier blockscout
```

4. Verify `VaultFactory`

4.1. Find `StakeVault` template address

```sh
$ cast call $VAULT_FACTORY_ADDRESS "vaultImplementation()(address)" --rpc-url $RPC_SN_SEPOLIA
```

4.2 Verify the vault factory

```sh
$ forge verify-contract \
--chain-id 1660990954 \
--num-of-optimizations 10000 \
--watch \
--compiler-version $COMPILER_VERSION \
$VAULT_FACTORY_CONTRACT_ADDRESS \
--rpc-url $RPC_SN_SEPOLIA \
--verifier-url $VERIFIER_URL_SN_SEPOLIA
--verifier blockscout \
--constructor-args $(cast abi-encode "constructor(address,address,address)" $OWNER_ADDRESS $STAKE_MANAGER_PROXY_CONTRACT_ADDRESS $VAULT_IMPL_ADDRESS)
```

5. Deploy Karma Token

```sh
$ MNEMONIC=$YOUR_MNEMONIC forge script script/DeployKarma.s.sol --rpc-url $RPC_SN_SEPOLIA --broadcast
```

6. Verify Karma Token

6.1 Get Karma implementation address from Karma Proxy:

```sh
$ cast storage $KARMA_PROXY_ADDRESS 0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc --rpc-url $RPC_SN_SEPOLIA
```

6.2 Verify Karma Implementation contract

```sh
$ forge verify-contract \
--chain-id 1660990954 \
--num-of-optimizations 10000 \
--watch \
--compiler-version $COMPILER_VERSION \
$KARMA_CONTRACT_ADDRESS \
--rpc-url $RPC_SN_SEPOLIA \
--verifier-url $VERIFIER_URL_SN_SEPOLIA
--verifier blockscout
```

7. Deploy KarmaNFT

```sh
$ MNEMONIC=$YOUR_MNEMONIC KARMA_ADDRESS=$KARMA_CONTRACT_ADDRESS forge script scripts/DeployKarmaNFT.s.sol --rpc-url $RPC_SN_SEPOLIA --broadcast
```

8. Verify Karma NFT

```sh
$ forge verify-contract \
--chain-id 1660990954 \
--num-of-optimizations 10000 \
--watch \
--compiler-version $COMPILER_VERSION \
$KARMA_NFT_CONTRACT_ADDRESS \
--rpc-url $RPC_SN_SEPOLIA \
--verifier-url $VERIFIER_URL_SN_SEPOLIA
--verifier blockscout \
--constructor-args $(cast abi-encode "constructor(address,address)" $KARMA_CONTRACT_ADDRESS $METADATA_GENERATOR_ADDRESS)
```

9. Set reward supplier in `StakeManager`

```sh
$ cast send $STAKE_MANAGER_PROXY_CONTRACT_ADDRESS "setRewardsSupplier(address)" $KARMA_PROXY_ADDRESS --rpc-url $RPC_SN_SEPOLIA --mnemonic $YOUR_MNEMONIC
```

10. Add `StakeManager` as reward distributor to `Karma` contract

```sh
$ cast send $KARMA_PROXY_CONTRACT_ADDRESS "addRewardDistributor(address)" $STAKE_MANAGER_PROXY_CONTRACT_ADDRESS --rpc-url $RPC_SN_SEPOLIA --mnemonic $YOUR_MNEMONIC
```

11. Whitelist `StakeVault` implementation in `StakeManager`

11.1 First, get the vault's codehash:

```sh
$ cast keccak $VAULT_BYTECODE
```

11.2 Then, set the resulting hash in the stake manager:

```sh
$ cast send $STAKE_MANAGER_PROXY_CONTRACT_ADDRESS "addTrustedCodehash(bytes32,bool)" $STAKE_VAULT_CODE_HASH true --rpc-url $RPC_SN_SEPOLIA --mnemonic $YOUR_MNEMONIC
```
