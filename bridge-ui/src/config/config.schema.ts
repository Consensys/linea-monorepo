import { isAddress } from "viem";
import { z } from "zod/v4-mini";

const chainConfigSchema = z.object({
  iconPath: z.string(),
  messageServiceAddress: z.string().check(
    z.refine((val) => isAddress(val), {
      message: "Invalid Ethereum address",
    }),
  ),
  tokenBridgeAddress: z.string().check(
    z.refine((val) => isAddress(val), {
      message: "Invalid Ethereum address",
    }),
  ),
  gasLimitSurplus: z.bigint().check(z.positive()),
  profitMargin: z.bigint().check(z.positive()),
  cctpDomain: z.number().check(z.gte(0), z.int()),
  cctpTokenMessengerV2Address: z.string().check(
    z.refine((val) => isAddress(val), {
      message: "Invalid Ethereum address",
    }),
  ),
  cctpMessageTransmitterV2Address: z.string().check(
    z.refine((val) => isAddress(val), {
      message: "Invalid Ethereum address",
    }),
  ),
});

export const configSchema = z.object({
  chains: z.record(z.string().check(z.regex(/^\d+$/)), chainConfigSchema),
  walletConnectId: z.string().check(z.minLength(1)),
  storage: z.object({
    minVersion: z.number().check(z.positive(), z.int()),
  }),
  // Feature toggle for CCTPV2 for USDC transfers
  isCctpEnabled: z.boolean(),
  infuraApiKey: z.string().check(z.minLength(1)),
  quickNodeApiKey: z.string().check(z.minLength(1)),
  dynamicEnvironmentId: z.string().check(z.minLength(1)),
  lifiApiKey: z.string().check(z.minLength(1)),
  lifiIntegrator: z.string().check(z.minLength(1)),
  onRamperApiKey: z.string().check(z.minLength(1)),
  layerswapApiKey: z.string().check(z.minLength(1)),
  tokenListUrls: z.object({
    mainnet: z.string().check(z.trim(), z.url()),
    sepolia: z.string().check(z.trim(), z.url()),
  }),
});

export type Config = z.infer<typeof configSchema>;
