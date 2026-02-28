import { defineConfig } from "hardhat/config";
import HardhatEthers from "@nomicfoundation/hardhat-ethers";
import HardhatEthersChaiMatchers from "@nomicfoundation/hardhat-ethers-chai-matchers";
// import HardhatFoundry from "@nomicfoundation/hardhat-foundry"; // Requires Foundry installation
import HardhatIgnitionEthers from "@nomicfoundation/hardhat-ignition-ethers";
import HardhatMocha from "@nomicfoundation/hardhat-mocha";
import HardhatNetworkHelpers from "@nomicfoundation/hardhat-network-helpers";
import HardhatTypechain from "@nomicfoundation/hardhat-typechain";
import HardhatVerify from "@nomicfoundation/hardhat-verify";
import * as dotenv from "dotenv";
import { getBlockchainNode, getL2BlockchainNode } from "./common.js";
import { SupportedChainIds } from "./common/supportedNetworks.js";
import { overrides } from "./hardhat_overrides.js";

// Task imports commented out - need migration to Hardhat v3 task format
// TODO: Migrate these tasks to use Hardhat v3 plugin task format

dotenv.config();

const BLOCKCHAIN_TIMEOUT = parseInt(process.env.BLOCKCHAIN_TIMEOUT_MS ?? "300000");
const EMPTY_HASH = "0x0000000000000000000000000000000000000000000000000000000000000000";

const blockchainNode = getBlockchainNode();
const l2BlockchainNode = getL2BlockchainNode();

const useViaIR = process.env.ENABLE_VIA_IR === "true";

export default defineConfig({
  plugins: [
    HardhatEthers,
    HardhatEthersChaiMatchers,
    // HardhatFoundry, // Requires Foundry installation
    HardhatIgnitionEthers,
    HardhatMocha,
    HardhatNetworkHelpers,
    HardhatTypechain,
    HardhatVerify,
  ],
  paths: {
    artifacts: "./build",
    sources: "./src",
    tests: "./test/hardhat",
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
      type: "edr-simulated",
      hardfork: "osaka",
      allowUnlimitedContractSize: true,
    },
    mainnet: {
      type: "http",
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      url: "https://mainnet.infura.io/v3/" + process.env.INFURA_API_KEY,
    },
    sepolia: {
      type: "http",
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      url: "https://sepolia.infura.io/v3/" + process.env.INFURA_API_KEY,
    },
    linea_mainnet: {
      type: "http",
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      url: "https://linea-mainnet.infura.io/v3/" + process.env.INFURA_API_KEY,
      chainId: 59144,
    },
    linea_sepolia: {
      type: "http",
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      url: "https://linea-sepolia.infura.io/v3/" + process.env.INFURA_API_KEY,
      chainId: SupportedChainIds.LINEA_SEPOLIA,
    },
    custom: {
      type: "http",
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      url: process.env.CUSTOM_RPC_URL || "http://localhost:8545",
    },
    zkevm_dev: {
      type: "http",
      gasPrice: 1322222229,
      url: blockchainNode,
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      timeout: BLOCKCHAIN_TIMEOUT,
      chainId: SupportedChainIds.LINEA_DEVNET,
    },
    l2: {
      type: "http",
      url: l2BlockchainNode || "http://localhost:8545",
      accounts: [process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH],
      allowUnlimitedContractSize: true,
    },
  },
  test: {
    mocha: {
      timeout: 20000,
    },
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
  typechain: {
    outDir: "typechain-types",
    target: "ethers-v6",
  },
});
