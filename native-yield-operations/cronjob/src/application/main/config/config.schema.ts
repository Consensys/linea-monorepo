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
    L1_RPC_URL: z.string().url(),
    STAKING_GRAPHQL_URL: z.string().url(),
    IPFS_BASE_URL: z.string().url(),

    LINEA_ROLLUP_ADDRESS: Address,
    LAZY_ORACLE_ADDRESS: Address,
    YIELD_MANAGER_ADDRESS: Address,
    LIDO_YIELD_PROVIDER_ADDRESS: Address,
    L2_YIELD_RECIPIENT: Address,

    // Optional API port (uncomment your apiOptions block to use)
    API_PORT: z.coerce.number().int().min(1024).max(49000),
  })
  .strict();

export type Config = z.infer<typeof configSchema>;
