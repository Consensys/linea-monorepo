import "@nomicfoundation/hardhat-toolbox";
import "@nomicfoundation/hardhat-ignition-ethers";
// import "@nomicfoundation/hardhat-foundry"; // Uncomment if Foundry is installed
import "@openzeppelin/hardhat-upgrades";
import "hardhat-storage-layout";
import "solidity-docgen";
import * as dotenv from "dotenv";
import { HardhatUserConfig } from "hardhat/config";
import { getBlockchainNode, getL2BlockchainNode } from "./common";
import { SupportedChainIds } from "./common/supportedNetworks";
import { overrides } from "./hardhat_overrides";
import "./scripts/operational/tasks/getCurrentFinalizedBlockNumberTask";
import "./scripts/operational/tasks/grantContractRolesTask";
import "./scripts/operational/tasks/renounceContractRolesTask";
import "./scripts/operational/tasks/setRateLimitTask";
import "./scripts/operational/tasks/setVerifierAddressTask";
import "./scripts/operational/tasks/setMessageServiceOnTokenBridgeTask";
import "./scripts/operational/yieldBoost/addLidoStVaultYieldProvider";
import "./scripts/operational/yieldBoost/prepareInitiateOssification";
import "./scripts/operational/yieldBoost/testing/addAndClaimMessage";
import "./scripts/operational/yieldBoost/testing/addAndClaimMessageForLST";
import "./scripts/operational/yieldBoost/testing/unstakePermissionless";

dotenv.config();

const BLOCKCHAIN_TIMEOUT = parseInt(process.env.BLOCKCHAIN_TIMEOUT_MS ?? "300000");
const EMPTY_HASH = "0x0000000000000000000000000000000000000000000000000000000000000000";

const blockchainNode = getBlockchainNode();
const l2BlockchainNode = getL2BlockchainNode();

const useViaIR = process.env.ENABLE_VIA_IR === "true";

const config: HardhatUserConfig = {
  paths: {
    artifacts: "./build",
    sources: "./src",
  },
  solidity: {
    compilers: [
      {
        version: "0.8.33",
        settings: {
          viaIR: useViaIR,
          optimizer: {
            enabled: true,
            runs: 10_000,
          },
          evmVersion: "osaka",
        },
      },
    ],
    overrides: overrides,
  },
  networks: {
    hardhat: {
      hardfork: "osaka",
      allowUnlimitedContractSize: true,
    },
    mainnet: {
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      url: "https://mainnet.infura.io/v3/" + process.env.INFURA_API_KEY,
    },
    sepolia: {
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      url: "https://sepolia.infura.io/v3/" + process.env.INFURA_API_KEY,
    },
    linea_mainnet: {
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      url: "https://linea-mainnet.infura.io/v3/" + process.env.INFURA_API_KEY,
      chainId: 59144,
    },
    linea_sepolia: {
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      url: "https://linea-sepolia.infura.io/v3/" + process.env.INFURA_API_KEY,
      chainId: SupportedChainIds.LINEA_SEPOLIA,
    },
    custom: {
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      url: process.env.CUSTOM_RPC_URL ? process.env.CUSTOM_RPC_URL : "",
    },
    zkevm_dev: {
      gasPrice: 1322222229,
      url: blockchainNode,
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      timeout: BLOCKCHAIN_TIMEOUT,
      chainId: SupportedChainIds.LINEA_DEVNET,
    },
    l2: {
      url: l2BlockchainNode ?? "",
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      allowUnlimitedContractSize: true,
    },
  },
  gasReporter: {
    enabled: !!process.env.REPORT_GAS,
  },
  mocha: {
    timeout: 20000,
  },
  etherscan: {
    apiKey: process.env.ETHERSCAN_API_KEY ?? "",
    customChains: [
      {
        network: "linea_sepolia",
        chainId: SupportedChainIds.LINEA_SEPOLIA,
        urls: {
          apiURL: `https://api.etherscan.io/v2/api?chainid=${SupportedChainIds.LINEA_SEPOLIA}`,
          browserURL: "https://sepolia.lineascan.build/",
        },
      },
      {
        network: "linea_mainnet",
        chainId: SupportedChainIds.LINEA,
        urls: {
          apiURL: `https://api.etherscan.io/v2/api?chainid=${SupportedChainIds.LINEA}`,
          browserURL: "https://lineascan.build/",
        },
      },
    ],
  },
  docgen: {
    exclude: [
      "_testing",
      "bridging/token/utils/StorageFiller39.sol",
      "bridging/token/CustomBridgedToken.sol",
      "governance/TimeLock.sol",
      "security/access/PermissionsManager.sol",
      "security/reentrancy/TransientStorageReentrancyGuardUpgradeable.sol",
      "tokens",
      "verifiers",
    ],
    pages: "files",
    outputDir: "docs/api/",
    pageExtension: ".mdx",
    templates: "docs/docgen-templates",
  },
};

export default config;
