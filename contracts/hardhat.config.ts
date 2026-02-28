import { defineConfig, configVariable } from "hardhat/config";
import hardhatToolboxMochaEthers from "@nomicfoundation/hardhat-toolbox-mocha-ethers";
import { SupportedChainIds } from "./common/supportedNetworks.js";
import { overrides } from "./hardhat_overrides.js";

const useViaIR = process.env.ENABLE_VIA_IR === "true";

export default defineConfig({
  plugins: [hardhatToolboxMochaEthers],
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
      chainType: "l1",
      hardfork: "osaka",
      allowUnlimitedContractSize: true,
    },
    mainnet: {
      type: "http",
      chainType: "l1",
      url: `https://mainnet.infura.io/v3/${configVariable("INFURA_API_KEY")}`,
      accounts: [configVariable("DEPLOYER_PRIVATE_KEY")],
    },
    sepolia: {
      type: "http",
      chainType: "l1",
      url: `https://sepolia.infura.io/v3/${configVariable("INFURA_API_KEY")}`,
      accounts: [configVariable("DEPLOYER_PRIVATE_KEY")],
    },
    linea_mainnet: {
      type: "http",
      chainType: "l1",
      url: `https://linea-mainnet.infura.io/v3/${configVariable("INFURA_API_KEY")}`,
      accounts: [configVariable("DEPLOYER_PRIVATE_KEY")],
      chainId: 59144,
    },
    linea_sepolia: {
      type: "http",
      chainType: "l1",
      url: `https://linea-sepolia.infura.io/v3/${configVariable("INFURA_API_KEY")}`,
      accounts: [configVariable("DEPLOYER_PRIVATE_KEY")],
      chainId: SupportedChainIds.LINEA_SEPOLIA,
    },
    custom: {
      type: "http",
      chainType: "l1",
      url: configVariable("CUSTOM_RPC_URL"),
      accounts: [configVariable("DEPLOYER_PRIVATE_KEY")],
    },
    zkevm_dev: {
      type: "http",
      chainType: "l1",
      url: configVariable("L1_RPC_URL"),
      accounts: [configVariable("DEPLOYER_PRIVATE_KEY")],
      chainId: SupportedChainIds.LINEA_DEVNET,
    },
    l2: {
      type: "http",
      chainType: "l1",
      url: configVariable("L2_RPC_URL"),
      accounts: [configVariable("DEPLOYER_PRIVATE_KEY")],
      allowUnlimitedContractSize: true,
    },
  },
  mocha: {
    timeout: 20000,
  },
  typechain: {
    outDir: "typechain-types",
  },
});
