// V2 plugin imports - commented out for V3 migration
// import "@nomicfoundation/hardhat-toolbox";
// import "@nomicfoundation/hardhat-foundry";
// import "@openzeppelin/hardhat-upgrades";
// import "hardhat-deploy";
// import "hardhat-storage-layout";
// import "hardhat-tracer"; // This plugin does not work with the latest hardhat version
// import "solidity-docgen";

// V3 imports
import { defineConfig } from "hardhat/config";
import hardhatToolboxMochaEthers from "@nomicfoundation/hardhat-toolbox-mocha-ethers";
import * as dotenv from "dotenv";

// Keep these imports but they may need updating for V3 compatibility
import { getBlockchainNode, getL2BlockchainNode } from "./common.js";
import { SupportedChainIds } from "./common/supportedNetworks.js";
import { overrides } from "./hardhat_overrides.js";

// V2 custom task imports - commented out for V3 migration (deferred)
// import "./scripts/operational/tasks/getCurrentFinalizedBlockNumberTask";
// import "./scripts/operational/tasks/grantContractRolesTask";
// import "./scripts/operational/tasks/renounceContractRolesTask";
// import "./scripts/operational/tasks/setRateLimitTask";
// import "./scripts/operational/tasks/setVerifierAddressTask";
// import "./scripts/operational/tasks/setMessageServiceOnTokenBridgeTask";

// TODO Later - Migrate to `npx hardhat keystore` and configVariable(), over process.env
dotenv.config();

const BLOCKCHAIN_TIMEOUT = parseInt(process.env.BLOCKCHAIN_TIMEOUT_MS ?? "300000");
const EMPTY_HASH = "0x0000000000000000000000000000000000000000000000000000000000000000";

const blockchainNode = getBlockchainNode();
const l2BlockchainNode = getL2BlockchainNode();

const useViaIR = process.env.ENABLE_VIA_IR === "true";

export default defineConfig({
  plugins: [hardhatToolboxMochaEthers],
  paths: {
    artifacts: "./build",
    sources: "./src",
    tests: "./test/hardhat", // Exclude foundry tests
  },
  solidity: {
    // NB: double check the autoupdate shell script version complies to the latest solidity version if you add a new one.
    /// @dev Please see the overrides file for a list of files not targetting the default EVM version of Prague.
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
    mainnet: {
      type: "http",
      accounts: [process.env.MAINNET_PRIVATE_KEY || EMPTY_HASH],
      url: "https://mainnet.infura.io/v3/" + process.env.INFURA_API_KEY,
    },
    sepolia: {
      type: "http",
      accounts: [process.env.SEPOLIA_PRIVATE_KEY || EMPTY_HASH],
      url: "https://sepolia.infura.io/v3/" + process.env.INFURA_API_KEY,
    },
    linea_mainnet: {
      type: "http",
      accounts: [process.env.LINEA_MAINNET_PRIVATE_KEY || EMPTY_HASH],
      url: "https://linea-mainnet.infura.io/v3/" + process.env.INFURA_API_KEY,
      chainId: 59144,
    },
    linea_sepolia: {
      type: "http",
      accounts: [process.env.LINEA_SEPOLIA_PRIVATE_KEY || EMPTY_HASH],
      url: "https://linea-sepolia.infura.io/v3/" + process.env.INFURA_API_KEY,
      chainId: SupportedChainIds.LINEA_SEPOLIA,
    },
    // Commented out networks with dynamic URLs - need to use  in V3
    custom: {
      type: "http",
      accounts: [process.env.CUSTOM_PRIVATE_KEY || EMPTY_HASH],
      url: process.env.CUSTOM_BLOCKCHAIN_URL ? process.env.CUSTOM_BLOCKCHAIN_URL : "https://example.com",
    },
    zkevm_dev: {
      type: "http",
      gasPrice: 1322222229,
      url: blockchainNode || "http://localhost:8545",
      accounts: [process.env.PRIVATE_KEY || EMPTY_HASH],
      timeout: BLOCKCHAIN_TIMEOUT,
      chainId: SupportedChainIds.LINEA_DEVNET,
    },
    l2: {
      type: "http",
      url: l2BlockchainNode ?? "https://example.com",
      accounts: [process.env.L2_PRIVATE_KEY || EMPTY_HASH],
    },
  },
  // gasReporter - removed (hardhat-gas-reporter specific, not available in V3)
  // gasReporter: {
  //   enabled: !!process.env.REPORT_GAS,
  // },
  // mocha: {
  //   timeout: 20000,
  // },
  verify: {
    etherscan: {
      apiKey: process.env.ETHERSCAN_API_KEY ?? "",
    },
  },
  // solidity-docgen not compatible with Hardhat V3 - https://github.com/OpenZeppelin/solidity-docgen/issues/471
  // TODO LATER - Trial forge doc as alternative docgen
  // docgen - commented out (solidity-docgen specific, not available in V3)
  // docgen: {
  //   exclude: [
  //     "_testing",
  //     "bridging/token/utils/StorageFiller39.sol",
  //     "bridging/token/CustomBridgedToken.sol",
  //     "governance/TimeLock.sol",
  //     "security/access/PermissionsManager.sol",
  //     "security/reentrancy/TransientStorageReentrancyGuardUpgradeable.sol",
  //     "tokens",
  //     "verifiers",
  //   ],
  //   pages: "files",
  //   outputDir: "docs/api/",
  //   // For compatibility with docs.linea.build
  //   pageExtension: ".mdx",
  //   templates: "docs/docgen-templates",
  // },
});
