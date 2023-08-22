import "@nomicfoundation/hardhat-toolbox";
import "@openzeppelin/hardhat-upgrades";
import * as dotenv from "dotenv";
import "hardhat-tracer";
import { HardhatUserConfig } from "hardhat/config";
import { MAX_GAS_LIMIT, getBlockchainNode, getContractOwnerPrivateKey, getL2BlockchainNode } from "./common";

dotenv.config();

const BLOCKCHAIN_TIMEOUT = process.env.BLOCKCHAIN_TIMEOUT_MS ? parseInt(process.env.BLOCKCHAIN_TIMEOUT_MS) : 300000;
const EMPTY_HASH = "0x0000000000000000000000000000000000000000000000000000000000000000"

const blockchainNode = getBlockchainNode();
const l2BlockchainNode = getL2BlockchainNode();

const ownerPrivateKey = getContractOwnerPrivateKey(
  "../node-data/test/keys/contract_owner.acc",
  "/node-data/test/keys/contract_owner.acc",
  // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
  process.env.PRIVATE_KEY!,
);

const accounts = [];

if (ownerPrivateKey) {
  accounts.push(ownerPrivateKey);
}

const config: HardhatUserConfig = {
  paths: {
    artifacts: "./build",
  },
  solidity: {
    compilers: [
      {
        version: "0.8.19",
        settings: {
          optimizer: {
            enabled: true,
            runs: 100000,
          },
        },
      },
      {
        version: "0.8.15",
        settings: {
          optimizer: {
            enabled: true,
            runs: 100000,
          },
        },
      },
    ],
  },
  networks: {
    mainnet: {
      accounts: [process.env.MAINNET_PRIVATE_KEY ?? EMPTY_HASH],
      url: "https://mainnet.infura.io/v3/" + process.env.INFURA_API_KEY
    },
    linea_mainnet: {
      accounts: [process.env.LINEA_MAINNET_PRIVATE_KEY ?? EMPTY_HASH],
      url: "https://linea-mainnet.infura.io/v3/" + process.env.INFURA_API_KEY
    },
    goerli: {
      accounts: [process.env.GOERLI_PRIVATE_KEY ?? EMPTY_HASH],
      url: "https://goerli.infura.io/v3/" + process.env.INFURA_API_KEY
    },
    linea_goerli: {
      accounts: [process.env.LINEA_GOERLI_PRIVATE_KEY ?? EMPTY_HASH],
      url: "https://linea-goerli.infura.io/v3/" + process.env.INFURA_API_KEY
    },
    besu: {
      url: blockchainNode,
      accounts,
      gasPrice: 0,
      gas: MAX_GAS_LIMIT,
      blockGasLimit: MAX_GAS_LIMIT,
      timeout: BLOCKCHAIN_TIMEOUT,
    },
    ganache: {
      url: "http://127.0.0.1:8545",
      accounts,
      gasPrice: 0,
      gas: MAX_GAS_LIMIT,
      blockGasLimit: MAX_GAS_LIMIT,
      timeout: BLOCKCHAIN_TIMEOUT,
    },
    zkevm_dev: {
      url: blockchainNode,
      accounts,
      timeout: BLOCKCHAIN_TIMEOUT,
    },
    ...(l2BlockchainNode
      ? {
          l2: {
            url: l2BlockchainNode,
            accounts,
            allowUnlimitedContractSize: true,
          },
        }
      : {}),
  },
  gasReporter: {
    enabled: !!process.env.REPORT_GAS,
  },
  mocha: {
    timeout: 20000,
  },
  etherscan: {
    apiKey: {
      mainnet: process.env.ETHERSCAN_API_KEY ?? "",
      goerli: process.env.ETHERSCAN_API_KEY ?? "",
      linea_goerli: process.env.LINEASCAN_API_KEY ?? "",
      linea_mainnet: process.env.LINEASCAN_API_KEY ?? "",
      // TODO Add for linea mainnet
    },
    customChains: [
      {
        network: "linea_goerli",
        chainId: 59140,
        urls: {
          apiURL: "https://api-goerli.lineascan.build/api",
          browserURL: "https://goerli.lineascan.build/",
        },
      },
      {
        network: "linea_mainnet",
        chainId: 59144,
        urls: {
          apiURL: "https://api.lineascan.build/api",
          browserURL: "https://lineascan.build/",
        },
      },
    ],
  },
};

export default config;
