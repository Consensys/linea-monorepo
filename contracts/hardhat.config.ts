import hardhatEthers from "@nomicfoundation/hardhat-ethers";
import hardhatEthersChaiMatchers from "@nomicfoundation/hardhat-ethers-chai-matchers";
import hardhatFoundry from "@nomicfoundation/hardhat-foundry";
import hardhatMocha from "@nomicfoundation/hardhat-mocha";
import hardhatNetworkHelpers from "@nomicfoundation/hardhat-network-helpers";
import hardhatTypechain from "@nomicfoundation/hardhat-typechain";
import hardhatVerify from "@nomicfoundation/hardhat-verify";
import hardhatUpgrades from "@openzeppelin/hardhat-upgrades";
import * as dotenv from "dotenv";
import { configVariable, defineConfig } from "hardhat/config";
import HardhatDeploy from "hardhat-deploy";

import { getBlockchainNode, getL2BlockchainNode } from "./common";
import { SupportedChainIds } from "./common/supportedNetworks";
import { overrides } from "./hardhat_overrides";
import { resolveDeployerAccounts } from "./scripts/hardhat/deployer-accounts";
import getCurrentFinalizedBlockNumberTask from "./scripts/operational/tasks/getCurrentFinalizedBlockNumberTask";
import grantContractRolesTask from "./scripts/operational/tasks/grantContractRolesTask";
import renounceContractRolesTask from "./scripts/operational/tasks/renounceContractRolesTask";
import setMessageServiceOnTokenBridgeTask from "./scripts/operational/tasks/setMessageServiceOnTokenBridgeTask";
import setRateLimitTask from "./scripts/operational/tasks/setRateLimitTask";
import setVerifierAddressTask from "./scripts/operational/tasks/setVerifierAddressTask";
import addLidoStVaultYieldProviderTask from "./scripts/operational/yieldBoost/addLidoStVaultYieldProvider";
import prepareInitiateOssificationTask from "./scripts/operational/yieldBoost/prepareInitiateOssification";
import addAndClaimMessageTask from "./scripts/operational/yieldBoost/testing/addAndClaimMessage";
import addAndClaimMessageForLSTTask from "./scripts/operational/yieldBoost/testing/addAndClaimMessageForLST";
import unstakePermissionlessTask from "./scripts/operational/yieldBoost/testing/unstakePermissionless";

dotenv.config();

const BLOCKCHAIN_TIMEOUT = parseInt(process.env.BLOCKCHAIN_TIMEOUT_MS ?? "300000");

/**
 * `HARDHAT_SIGNER_UI=true` -> no local keys (browser wallet via signer-ui bridge).
 * If `DEPLOYER_PRIVATE_KEY` is unset -> `[]` so `hardhat build`, `clean`, etc. work without secrets
 * (same as typical Hardhat: RPC-only until you sign). Deploy/sign on a live network then needs a key
 * or `HARDHAT_SIGNER_UI=true`. If a key *is* set, it must be valid non-zero hex (all-zero breaks
 * LocalAccountsProvider / @ethereumjs/util).
 */
function deployerAccounts(): string[] {
  return resolveDeployerAccounts();
}

function infuraUrl(network: string): string {
  return `https://${network}.infura.io/v3/${process.env.INFURA_API_KEY ?? ""}`;
}

const blockchainNode = getBlockchainNode();
const l2BlockchainNode = getL2BlockchainNode();

const useViaIR = process.env.ENABLE_VIA_IR === "true";

const osakaCompiler = {
  version: "0.8.33",
  settings: {
    viaIR: useViaIR,
    optimizer: {
      enabled: true,
      runs: 10_000,
    },
    evmVersion: "osaka" as const,
  },
};

const hardhatNetwork = {
  type: "edr-simulated" as const,
  chainType: "l1" as const,
  hardfork: "osaka",
  // NB: Remove when ready for Deploying to Mainnet
  allowUnlimitedContractSize: true,
};

const operationalTasks = [
  getCurrentFinalizedBlockNumberTask,
  grantContractRolesTask,
  renounceContractRolesTask,
  setMessageServiceOnTokenBridgeTask,
  setRateLimitTask,
  setVerifierAddressTask,
  addLidoStVaultYieldProviderTask,
  prepareInitiateOssificationTask,
  addAndClaimMessageTask,
  addAndClaimMessageForLSTTask,
  unstakePermissionlessTask,
];

export default defineConfig({
  plugins: [
    hardhatEthers,
    hardhatEthersChaiMatchers,
    hardhatFoundry,
    hardhatMocha,
    hardhatNetworkHelpers,
    hardhatTypechain,
    hardhatVerify,
    hardhatUpgrades,
    HardhatDeploy,
  ],
  tasks: operationalTasks,
  paths: {
    artifacts: "./build",
    sources: "./src",
    tests: {
      mocha: "./test/hardhat",
    },
  },
  solidity: {
    // NB: double check the autoupdate shell script version complies to the latest solidity version if you add a new one.
    /// @dev Please see the overrides file for a list of files not targetting the default EVM version of Prague.
    profiles: {
      default: {
        compilers: [osakaCompiler],
        overrides,
      },
      production: {
        compilers: [osakaCompiler],
        overrides,
      },
    },
  },
  typechain: {
    outDir: "typechain-types",
  },
  networks: {
    default: hardhatNetwork,
    hardhat: hardhatNetwork,
    node: hardhatNetwork,
    mainnet: {
      type: "http",
      chainType: "l1",
      accounts: deployerAccounts(),
      url: infuraUrl("mainnet"),
    },
    sepolia: {
      type: "http",
      chainType: "l1",
      accounts: deployerAccounts(),
      url: infuraUrl("sepolia"),
    },
    hoodi: {
      type: "http",
      chainType: "l1",
      accounts: deployerAccounts(),
      url: infuraUrl("hoodi"),
      chainId: SupportedChainIds.HOODI,
    },
    linea_mainnet: {
      type: "http",
      chainType: "generic",
      accounts: deployerAccounts(),
      url: infuraUrl("linea-mainnet"),
      chainId: 59144,
    },
    linea_sepolia: {
      type: "http",
      chainType: "generic",
      accounts: deployerAccounts(),
      url: infuraUrl("linea-sepolia"),
      chainId: SupportedChainIds.LINEA_SEPOLIA,
    },
    custom: {
      type: "http",
      chainType: "generic",
      accounts: deployerAccounts(),
      url: process.env.CUSTOM_RPC_URL ?? configVariable("CUSTOM_RPC_URL"),
    },
    zkevm_dev: {
      type: "http",
      chainType: "generic",
      gasPrice: 1322222229,
      url: blockchainNode,
      accounts: deployerAccounts(),
      timeout: BLOCKCHAIN_TIMEOUT,
      // docker L1 is 31648428 (docker/config/l1-node/el/genesis.json);
      // hosted devnet (e.g. rpc.devnet.linea.build) uses 59139.
      // Set ZKEVM_DEV_CHAIN_ID to enforce chain-ID validation and prevent wrong-chain deployments.
      // Omitting it disables validation and preserves flexibility across environments.
      ...(process.env.ZKEVM_DEV_CHAIN_ID ? { chainId: parseInt(process.env.ZKEVM_DEV_CHAIN_ID, 10) } : {}),
    },
    l2: {
      type: "http",
      chainType: "generic",
      url: l2BlockchainNode ?? configVariable("L2_BLOCKCHAIN_NODE"),
      accounts: deployerAccounts(),
    },
  },
  verify: {
    etherscan: {
      apiKey: process.env.ETHERSCAN_API_KEY ?? "",
    },
  },
  test: {
    mocha: {
      timeout: 20000,
    },
  },
});
