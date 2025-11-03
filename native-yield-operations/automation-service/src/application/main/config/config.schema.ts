// src/config.ts
import { z } from "zod";
import { isAddress, getAddress, isHex } from "viem";

/** Reusable EVM address schema: validates then normalizes to checksummed form */
const Address = z
  .string()
  .refine((v) => isAddress(v), { message: "Invalid Ethereum address" })
  .transform((v) => getAddress(v)); // checksum/normalize

const Hex = z.string().refine((v) => isHex(v), { message: "Invalid Hex" });

export const configSchema = z
  .object({
    // Datasource URLs
    CHAIN_ID: z.coerce.number().int().positive(),
    L1_RPC_URL: z.string().url(),
    BEACON_CHAIN_RPC_URL: z.string().url(),
    STAKING_GRAPHQL_URL: z.string().url(),
    IPFS_BASE_URL: z.string().url(),
    // Consensys Staking OAuth2 Token Endpoint
    CONSENSYS_STAKING_OAUTH2_TOKEN_ENDPOINT: z.string().url(),
    CONSENSYS_STAKING_OAUTH2_CLIENT_ID: z.string().min(1),
    CONSENSYS_STAKING_OAUTH2_CLIENT_SECRET: z.string().min(1),
    CONSENSYS_STAKING_OAUTH2_AUDIENCE: z.string().min(1),
    // L1 contract addresses
    LINEA_ROLLUP_ADDRESS: Address,
    LAZY_ORACLE_ADDRESS: Address,
    VAULT_HUB_ADDRESS: Address,
    YIELD_MANAGER_ADDRESS: Address,
    LIDO_YIELD_PROVIDER_ADDRESS: Address,
    // L2 contract addresses
    L2_YIELD_RECIPIENT: Address,
    // Timing intervals
    TRIGGER_EVENT_POLL_INTERVAL_MS: z.coerce.number().int().positive(),
    TRIGGER_MAX_INACTION_MS: z.coerce.number().int().positive(),
    CONTRACT_READ_RETRY_TIME_MS: z.coerce.number().int().positive(),
    // Tolerance band for changes around the target threshold reserve, no rebalance will be done unless exceed this band.
    // Bps is multiplied by YieldManager.totalSystemBalance().
    REBALANCE_TOLERANCE_BPS: z.coerce.number().int().positive().max(10000),
    // Unstake params
    MAX_VALIDATOR_WITHDRAWAL_REQUESTS_PER_TRANSACTION: z.coerce.number().int().positive(),
    // Minimum withdrawal threshold — no withdrawal occurs if available amount < threshold.
    MIN_WITHDRAWAL_THRESHOLD_ETH: z
      .union([z.string(), z.number(), z.bigint()])
      .transform((val) => BigInt(val))
      .refine((v) => v >= 0n, { message: "Must be nonnegative" }),
    // Web3Signer
    WEB3SIGNER_URL: z.string().url(),
    // Accept either an Ethereum address (20 bytes) OR a secp256k1 pubkey (33/65 bytes).
    // If you only want addresses, replace with `Address`.
    WEB3SIGNER_PUBLIC_KEY: Hex.refine(
      (v) =>
        /^(?:0x)?[a-fA-F0-9]{128}$/.test(v) || // uncompressed pubkey (64 bytes, without ...04-prefix, optional 0x prefix)
        "Expected secp256k1 public key (uncompressed, without 0x04 prefix).",
    ),
    WEB3SIGNER_KEYSTORE_PATH: z.string().min(1),
    WEB3SIGNER_KEYSTORE_PASSPHRASE: z.string().min(1),
    WEB3SIGNER_TRUSTSTORE_PATH: z.string().min(1),
    WEB3SIGNER_TRUSTSTORE_PASSPHRASE: z.string().min(1),
    WEB3SIGNER_TLS_ENABLED: z.coerce.boolean(),
    // API port
    API_PORT: z.coerce.number().int().min(1024).max(49000),
  })
  .strict();

export type FlattenedConfigSchema = z.infer<typeof configSchema>;
