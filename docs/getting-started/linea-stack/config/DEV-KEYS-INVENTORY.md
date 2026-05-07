# Dev keys inventory — DO NOT REUSE

> Every file listed here is a private key checked into a public repository.
> Anyone in the world can spend funds these keys control. They exist solely
> to make `docker compose up` work on a fresh laptop. Never reuse, copy,
> or derive from them for any non-local network.

| Path | Format | Used by | Public knowledge? |
|------|--------|---------|-------------------|
| `config/l1/el/besu.key`                                    | raw hex SECP256K1 | L1 Besu node identity | yes |
| `config/l1/cl/teku.key`                                    | raw hex SECP256K1 | L1 Teku node identity | yes |
| `config/l1/cl/teku-keys/0x*.json`                          | EIP-2335 keystore | L1 Teku validator keys | yes |
| `config/l1/cl/teku-secrets/0x*.txt`                        | password files     | L1 Teku validator passwords | yes |
| `config/l1/genesis-generator/mnemonics.yaml`               | BIP-39 mnemonic   | L1 genesis validator + funded accounts | yes |
| `config/l2/sequencer/key`                                  | raw hex SECP256K1 | Linea Besu sequencer node identity | yes |
| `config/l2/maru/private-key`                               | raw hex SECP256K1 | Maru consensus signer | yes |
| `config/web3signer/key-files/anchoring-signer.yaml`        | YAML wrapping raw key | L2 message anchoring | yes |
| `config/web3signer/key-files/data-submission-signer.yaml`  | YAML wrapping raw key | L1 blob/calldata submission | yes |
| `config/web3signer/key-files/finalization-signer.yaml`     | YAML wrapping raw key | L1 aggregation finalization | yes |
| `config/web3signer/key-files/liveness-signer.yaml`         | YAML wrapping raw key | L1 liveness heartbeat | yes |
| `config/web3signer/tls-files/web3signer-keystore.p12`      | PKCS#12 keystore | Web3signer TLS server cert | yes |
| `config/web3signer/tls-files/web3signer-keystore-password.txt` | plaintext password | Web3signer TLS keystore password | yes |
| `config/l2/postman/env` (`L1_SIGNER_PRIVATE_KEY`, `L2_SIGNER_PRIVATE_KEY`) | hex strings inline | Postman L1 + L2 signers | yes |

## Replacing keys for any non-loopback deployment

Before running this stack on a host reachable from outside `localhost`:

1. Regenerate every web3signer raw key (use `eth-account` or any standard keygen).
2. Rebuild the web3signer mTLS keystore (`openssl pkcs12 -export …`) and update `known-clients.txt` with the new client cert fingerprint.
3. Regenerate the maru and sequencer node keys — these are P2P identity keys, so the bootnode enode strings in `docker-compose.yml` (currently `enode://14408801…@sequencer:30303`) will need to be regenerated to match.
4. Replace the L1 EL/CL identity keys (`besu.key`, `teku.key`).
5. Replace the L1 Teku validator keys + secrets (4 keystore/password pairs).
6. Replace the L1 genesis generator mnemonic, then rebuild L1 genesis.
7. Update the postman env file's signer keys.
8. Re-run `deploy-contracts.sh` so the deployer-derived addresses match the new keys.

If you can't tick all eight, keep the stack on `localhost`.
