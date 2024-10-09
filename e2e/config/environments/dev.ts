import { ethers } from "ethers";
import { TestnetAccountManager } from "../accounts/testnet-account-manager";
import { Config } from "../type";
import Account from "../accounts/account";

const L1_RPC_URL = new URL(`https://sepolia.infura.io/v3/${process.env.INFURA_PROJECT_ID}`);
const L2_RPC_URL = new URL("https://rpc.devnet.linea.build");

const L1_WHALE_ACCOUNTS: Account[] = [];
const L2_WHALE_ACCOUNTS: Account[] = [];

const config: Config = {
  L1: {
    rpcUrl: L1_RPC_URL,
    chainId: 11155111,
    lineaRollupAddress: "0x2A5CDCfc38856e2590E9Bd32F54Fa348e5De5f48",
    accountManager: new TestnetAccountManager(new ethers.JsonRpcProvider(L1_RPC_URL.toString()), L1_WHALE_ACCOUNTS),
    dummyContractAddress: "",
  },
  L2: {
    rpcUrl: L2_RPC_URL,
    chainId: 59139,
    l2MessageServiceAddress: "0x33bf916373159A8c1b54b025202517BfDbB7863D",
    accountManager: new TestnetAccountManager(new ethers.JsonRpcProvider(L2_RPC_URL.toString()), L2_WHALE_ACCOUNTS),
    dummyContractAddress: "",
  },
};

export default config;
