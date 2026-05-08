# Dev keys inventory — DO NOT REUSE

> Every file listed below is a private key checked into a public repository.
> Anyone in the world can spend funds these keys control. They exist solely
> to make `docker compose up` work on a fresh laptop. Never reuse, copy, or
> derive from them for any non-loopback or non-Sepolia network.

The Sepolia migration (Phases 1–3) split keys into two groups:

- **L2 dev keys (still committed)** — we own the L2 genesis; these are
  intentionally checked in. Listed below.
- **L1 keys (user-supplied/derived at boot)** — your `L1_DEPLOYER_PRIVATE_KEY`
  from `.env` deploys and admins the L1 contracts. `account-setup` derives
  separate L1 blob and finalisation submitter keys from it, records their
  addresses/pubkeys, and renders web3signer YAMLs at boot. deploy-contracts
  grants/funds those derived addresses. Nothing about your L1 key is committed.

## L2 dev keys (committed)

| Path | Format | Used by | Public knowledge? |
|------|--------|---------|-------------------|
| `config/l2/sequencer/key`                                  | raw hex SECP256K1 | Linea Besu sequencer node identity | yes |
| `config/l2/maru/private-key`                               | raw hex SECP256K1 | Maru consensus signer | yes |
| `config/l2/postman/env` (`L2_SIGNER_PRIVATE_KEY`)          | hex string inline | Postman L2 EOA | yes |
| `genesis-besu.json.template` (multiple `privateKey` keys)  | hex strings inline | L2 genesis pre-funded EOAs (deployer, accounts 2-21, security council, message anchorer, etc.) | yes |
| `scripts/deploy-contracts.sh` (`L2_DEPLOYER_PRIVATE_KEY` default) | hex string inline | L2 contract deployer | yes |
| `scripts/account-setup.sh` (`L2_LIVENESS_SIGNER_PRIVATE_KEY` default) | hex string inline | L2 sequencer-liveness signer | yes |

## TLS material (committed; mTLS only — not Sepolia-funds-controlling)

| Path | Format | Used by |
|------|--------|---------|
| `config/web3signer/tls-files/web3signer-keystore.p12`           | PKCS#12 keystore       | Web3signer mTLS server cert |
| `config/web3signer/tls-files/web3signer-keystore-password.txt`  | plaintext password     | Web3signer keystore password |
| `config/web3signer/tls-files/known-clients.txt`                 | client cert fingerprint list | Web3signer mTLS client allow-list |
| `config/l2/coordinator/tls-files/coordinator-client-keystore.p12` | PKCS#12 keystore       | Coordinator mTLS client cert (talks to web3signer) |
| `config/l2/sequencer/tls-files/sequencer_client_keystore.p12`   | PKCS#12 keystore       | Sequencer mTLS client cert |

## Runtime-rendered keys (NOT committed)

`account-setup` writes these to a docker volume at boot from your `.env`:

- `/shared/web3signer-keys/anchoring-signer.yaml`        — L2 message anchoring (the `L2_MESSAGE_ANCHORING_PRIVATE_KEY` dev key)
- `/shared/web3signer-keys/data-submission-signer.yaml`  — L1 blob/data submission (derived from your L1 deployer key)
- `/shared/web3signer-keys/finalization-signer.yaml`     — L1 aggregation finalization (derived from your L1 deployer key)
- `/shared/web3signer-keys/liveness-signer.yaml`         — L2 sequencer liveness (the `L2_LIVENESS_SIGNER_PRIVATE_KEY` dev key)

`docker compose down -v` wipes them; the next `up` re-renders.

## Replacing keys for any non-loopback / non-Sepolia deployment

Before running this stack on a host that targets anything other than Sepolia
testnet (e.g., a real testnet you control, or — emphatically NOT recommended
without a full audit — mainnet):

1. **L2 keys**: regenerate `config/l2/sequencer/key` (the sequencer P2P identity
   — and update the bootnode `enode://…` in `docker-compose.yml`'s l2-node-besu
   service to match the new pubkey), `config/l2/maru/private-key`, and every
   pre-funded address in `genesis-besu.json.template`. Update `L2_SIGNER_PRIVATE_KEY`
   in `config/l2/postman/env`. Update `L2_DEPLOYER_PRIVATE_KEY` and
   `L2_MESSAGE_ANCHORING_PRIVATE_KEY`, and `L2_LIVENESS_SIGNER_PRIVATE_KEY`
   defaults in the scripts.
2. **TLS material**: regenerate the four `*.p12` keystores and the
   `known-clients.txt` fingerprint list. Update the keystore password.
3. **L1 keys**: already user-supplied via `.env` — just use a fresh deployer key.
4. Re-run `account-setup` so derived addresses + rendered web3signer YAMLs
   match the new keys.

If you can't tick all four, keep this stack on Sepolia.
