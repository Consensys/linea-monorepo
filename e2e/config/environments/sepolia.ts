import { ethers } from "ethers";
import { TestnetAccountManager } from "../accounts/testnet-account-manager";
import Account from "../accounts/account";
import { Config } from "../type";

const L1_RPC_URL = new URL(`https://sepolia.infura.io/v3/${process.env.INFURA_PROJECT_ID}`);
const L2_RPC_URL = new URL(`https://linea-sepolia.infura.io/v3/${process.env.INFURA_PROJECT_ID}`);

const L1_WHALE_ACCOUNTS: Account[] = [];
const L2_WHALE_ACCOUNTS: Account[] = [];

const config: Config = {
  L1: {
    rpcUrl: L1_RPC_URL,
    chainId: 11155111,
    lineaRollupAddress: "0xB218f8A4Bc926cF1cA7b3423c154a0D627Bdb7E5",
    accountManager: new TestnetAccountManager(new ethers.JsonRpcProvider(L1_RPC_URL.toString()), L1_WHALE_ACCOUNTS),
    dummyContractAddress: "",
  },
  L2: {
    rpcUrl: L2_RPC_URL,
    chainId: 59141,
    l2MessageServiceAddress: "0x971e727e956690b9957be6d51Ec16E73AcAC83A7",
    accountManager: new TestnetAccountManager(new ethers.JsonRpcProvider(L2_RPC_URL.toString()), L2_WHALE_ACCOUNTS),
    dummyContractAddress: "",
  },
};

export default config;
