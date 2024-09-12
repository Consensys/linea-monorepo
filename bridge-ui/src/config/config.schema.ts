import Joi from "joi";

export const configSchema = Joi.object({
  history: Joi.object({
    totalBlocksToParse: Joi.number().required(),
    blocksPerLoop: Joi.number().required(),
  }),
  networks: Joi.object({
    MAINNET: Joi.object({
      L1: Joi.object({
        name: Joi.string().required(),
        iconPath: Joi.string().required(),
        chainId: Joi.number().required(),
        messageServiceAddress: Joi.string().required(),
        tokenBridgeAddress: Joi.string().required(),
        usdcBridgeAddress: Joi.string().required(),
      }),
      L2: Joi.object({
        name: Joi.string().required(),
        iconPath: Joi.string().required(),
        chainId: Joi.number().required(),
        messageServiceAddress: Joi.string().required(),
        tokenBridgeAddress: Joi.string().required(),
        usdcBridgeAddress: Joi.string().required(),
      }),
      gasEstimated: Joi.required(),
      gasLimitSurplus: Joi.required(),
      profitMargin: Joi.required(),
    }),
    SEPOLIA: Joi.object({
      L1: Joi.object({
        name: Joi.string().required(),
        iconPath: Joi.string().required(),
        chainId: Joi.number().required(),
        messageServiceAddress: Joi.string().required(),
        tokenBridgeAddress: Joi.string().required(),
        usdcBridgeAddress: Joi.string().required(),
      }),
      L2: Joi.object({
        name: Joi.string().required(),
        iconPath: Joi.string().required(),
        chainId: Joi.number().required(),
        messageServiceAddress: Joi.string().required(),
        tokenBridgeAddress: Joi.string().required(),
        usdcBridgeAddress: Joi.string().required(),
      }),
      gasEstimated: Joi.required(),
      gasLimitSurplus: Joi.required(),
      profitMargin: Joi.required(),
    }),
  }),
  walletConnectId: Joi.string().disallow("").required(),
  storage: Joi.object({
    minVersion: Joi.string().required(),
  }),
});
