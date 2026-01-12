const useViaIR = process.env.USE_VIA_IR === "true";

const lineaOverride = {
  version: "0.8.33",
  settings: {
    evmVersion: "osaka",
    optimizer: {
      enabled: true,
      runs: 10_000,
    },
  },
};

const lineaOverridePaths = [
  "src/messaging/l2/L2MessageService.sol",
  "src/operational/LineaSequencerUptimeFeed.sol",
  "src/bridging/token/TokenBridge.sol",
  "src/bridging/token/BridgedToken.sol",
  "src/tokens/LineaVoyageXP.sol",
  "src/tokens/LineaSurgeXP.sol",
  "src/tokens/TokenMintingRateLimiter.sol",
  "src/_testing/mocks/bridging/MockMessageService.sol",
  "src/_testing/mocks/bridging/MockMessageServiceV2.sol",
  "src/_testing/mocks/bridging/MockTokenBridge.sol",
  "src/_testing/mocks/bridging/TestTokenBridge.sol",
  "src/_testing/mocks/bridging/UpgradedBridgedToken.sol",
  "src/_testing/unit/messaging/TestL2MessageManager.sol",
  "src/_testing/unit/messaging/TestL2MessageService.sol",
  "src/predeploy/UpgradeableConsolidationQueuePredeploy.sol",
  "src/predeploy/UpgradeableWithdrawalQueuePredeploy.sol",
  "src/predeploy/UpgradeableBeaconChainDepositPredeploy.sol",
  "src/operational/RollupRevenueVault.sol",
];

const specificOverrides: Record<string, object> = {
  "contracts/yield/YieldManager.sol": {
    version: "0.8.33",
    settings: {
      viaIR: useViaIR,
      optimizer: {
        enabled: true,
        runs: 1500,
      },
      evmVersion: "osaka",
    },
  },
  "contracts/test-contracts/TestLineaRollup.sol": {
    version: "0.8.33",
    settings: {
      viaIR: useViaIR,
      optimizer: {
        enabled: true,
        runs: 1000,
      },
      evmVersion: "osaka",
    },
  },
  "contracts/LineaRollup.sol": {
    version: "0.8.33",
    settings: {
      viaIR: useViaIR,
      optimizer: {
        enabled: true,
        runs: 9000,
      },
      evmVersion: "osaka",
    },
  },
};

export const overrides = {
  ...Object.fromEntries(lineaOverridePaths.map((path) => [path, lineaOverride])),
  ...specificOverrides,
};
