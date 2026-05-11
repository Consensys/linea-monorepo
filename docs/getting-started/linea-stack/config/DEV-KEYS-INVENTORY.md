# Dev keys inventory — DO NOT REUSE

> Every private key listed in the committed section is public knowledge. These
> keys exist only to make the local quickstart convenient. Never reuse, copy, or
> derive from them for any non-loopback or non-Sepolia network.

The quickstart now has three key classes:

- **User-supplied L1 deployer** — `L1_DEPLOYER_PRIVATE_KEY` in `.env`. It must
  be funded on Sepolia and is the only user input secret.
- **Generated runtime keys** — `account-setup` creates L1 blob/data-submission,
  L1 finalization, L1 postman, L2 deployer, L2 anchorer, and L2 postman keys at
  first boot, then persists them in the Docker shared volume for retries.
- **Committed local-dev keys** — static node identity and consensus keys that
  remain in the repo and are listed below.

## Committed local-dev keys

| Path | Format | Used by | Public knowledge? |
|------|--------|---------|-------------------|
| `config/l2/sequencer/key` | raw hex SECP256K1 | Linea Besu sequencer node identity | yes |
| `config/l2/maru/private-key` | raw hex SECP256K1 | Maru consensus signer | yes |

## TLS material (committed; mTLS only — not Sepolia-funds-controlling)

| Path | Format | Used by |
|------|--------|---------|
| `config/web3signer/tls-files/web3signer-keystore.p12` | PKCS#12 keystore | Web3signer mTLS server cert |
| `config/web3signer/tls-files/web3signer-keystore-password.txt` | plaintext password | Web3signer keystore password |
| `config/web3signer/tls-files/known-clients.txt` | client cert fingerprint list | Web3signer mTLS client allow-list |
| `config/l2/coordinator/tls-files/coordinator-client-keystore.p12` | PKCS#12 keystore | Coordinator mTLS client cert |
| `config/l2/sequencer/tls-files/sequencer_client_keystore.p12` | PKCS#12 keystore | Sequencer mTLS client cert |

## Runtime-generated keys (not committed)

`account-setup` writes these into the `linea-shared-config` Docker volume:

- `/shared/runtime-keys.env`
  - `L1_BLOB_SUBMITTER_PRIVATE_KEY`
  - `L1_FINALIZATION_SUBMITTER_PRIVATE_KEY`
  - `L1_POSTMAN_PRIVATE_KEY`
  - `L2_DEPLOYER_PRIVATE_KEY`
  - `L2_MESSAGE_ANCHORING_PRIVATE_KEY`
  - `L2_POSTMAN_PRIVATE_KEY`
- `/shared/web3signer-keys/anchoring-signer.yaml` — generated L2 message anchoring key
- `/shared/web3signer-keys/data-submission-signer.yaml` — generated L1 blob/data-submission key
- `/shared/web3signer-keys/finalization-signer.yaml` — generated L1 aggregation/finalization key

`docker compose down -v` wipes these volume files; the next boot generates a
fresh runtime key set.

## Replacing keys for any non-loopback / non-Sepolia deployment

Before running this stack against anything other than Sepolia testnet:

1. Use a fresh, funded `L1_DEPLOYER_PRIVATE_KEY`.
2. Regenerate `config/l2/sequencer/key` and update the l2-node-besu bootnode
   `enode://...` in `docker-compose.yml` to match the new pubkey.
3. Regenerate `config/l2/maru/private-key`.
4. Regenerate the committed TLS material if the deployment is not purely local.
5. Start from an empty Docker volume so `account-setup` generates fresh runtime
   keys and derived addresses.

If you cannot tick these off, keep this stack on Sepolia.
