// src/config.ts
import { z } from "zod";
import { isAddress, getAddress } from "viem";

/** Reusable EVM address schema: validates then normalizes to checksummed form */
const Address = z
  .string()
  .refine((v) => isAddress(v), { message: "Invalid Ethereum address" })
  .transform((v) => getAddress(v)); // checksum/normalize

export const configSchema = z
  .object({
    // Datasource URLs
    L1_RPC_URL: z.string().url(),
    STAKING_GRAPHQL_URL: z.string().url(),
    IPFS_BASE_URL: z.string().url(),
    // L1 contract addresses
    LINEA_ROLLUP_ADDRESS: Address,
    LAZY_ORACLE_ADDRESS: Address,
    YIELD_MANAGER_ADDRESS: Address,
    LIDO_YIELD_PROVIDER_ADDRESS: Address,
    // L2 contract addresses
    L2_YIELD_RECIPIENT: Address,
    // Timing intervals
    TRIGGER_EVENT_POLLING_TIME_SECONDS: z.coerce.number().int().positive(),
    TRIGGER_MAX_INACTION_TIMEOUT_SECONDS: z.coerce.number().int().positive(),
    CONTRACT_READ_RETRY_TIME_SECONDS: z.coerce.number().int().positive(),                        
    // API port
    API_PORT: z.coerce.number().int().min(1024).max(49000),
  })
  .strict();

export type FlattenedConfigSchema = z.infer<typeof configSchema>;
