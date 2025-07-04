import { formatEther, formatUnits, parseEther } from "viem";
import "dotenv/config";

export const METAMASK_SEED_PHRASE = process.env.E2E_TEST_SEED_PHRASE;
export const METAMASK_PASSWORD = process.env.E2E_TEST_WALLET_PASSWORD;

export const LOCAL_L1_NETWORK = {
  name: "L1",
  rpcUrl: `http://localhost:8445`,
  chainId: 31648428,
  symbol: "ETH",
  blockExplorerUrl: "https://etherscan.io",
};

export const LOCAL_L2_NETWORK = {
  name: "L2",
  rpcUrl: `http://localhost:9045`,
  chainId: 1337,
  symbol: "ETH",
  blockExplorerUrl: "https://lineascan.build",
};

export const TEST_URL = "http://localhost:3000/";
export const WEI_AMOUNT = formatEther(parseEther("1"));
// Must be > minimum CCTP fee
export const USDC_AMOUNT = formatUnits(10n, 6).toString();
export const ETH_SYMBOL = "ETH";
export const USDC_SYMBOL = "USDC";

export const POLLING_INTERVAL = 250;
export const PAGE_TIMEOUT = 10000;
