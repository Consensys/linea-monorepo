import { ethers } from "ethers";
import * as dotenv from "dotenv";
dotenv.config();

export const TRANSACTION_CALLDATA_LIMIT = 59000;
export const L1_RPC_URL = "https://goerli.infura.io/v3/" + process.env.UAT_L1_RPC_KEY;
export const L2_RPC_URL = "https://linea-goerli.infura.io/v3/" + process.env.UAT_L2_RPC_KEY;
export const CHAIN_ID = 59140;

export function getL1Provider() {
  return new ethers.providers.JsonRpcProvider(L1_RPC_URL);
}

export function getL2Provider() {
  return new ethers.providers.JsonRpcProvider(L2_RPC_URL);
}

export const DEPLOYER_ACCOUNT_PRIVATE_KEY = process.env.UAT_DEPLOYER_ACCOUNT_PRIVATE_KEY ?? "";

export const INITIAL_WITHDRAW_LIMIT = ethers.utils.parseEther("5");

export const ACCOUNT_0 = "0x174634fbF3d3e243543c6F22102837A113DF9005";
export const ACCOUNT_0_PRIVATE_KEY = process.env.UAT_ACCOUNT_0_PRIVATE_KEY ?? "";

export const ACCOUNT_1 = "";

export const OPERATOR_0 = "0xA2689249FAeAf6D84b6087A2970a20432b86A53e";

export const LINEA_ROLLUP_INITIAL_STATE_ROOT_HASH =
  "0x0000000000000000000000000000000000000000000000000000000000000000";
export const LINEA_ROLLUP_INITIAL_L2_BLOCK_NR = "123456";
export const LINEA_ROLLUP_SECURITY_COUNCIL = OPERATOR_0;
export const LINEA_ROLLUP_OPERATORS = [OPERATOR_0];
export const LINEA_ROLLUP_RATE_LIMIT_PERIOD = "86400"; //24Hours in seconds
export const LINEA_ROLLUP_RATE_LIMIT_AMOUNT = "1000000000000000000000"; //1000ETH

export const LINEA_ROLLUP_CONTRACT_ADDRESS = "0x70BaD09280FD342D02fe64119779BC1f0791BAC2";
export const MESSAGE_SERVICE_ADDRESS = "0xC499a572640B64eA1C8c194c43Bc3E19940719dC";
export const DUMMY_CONTRACT_ADDRESS = "0x91614d09f5A8c87E28ab79D5eC48164DA0988319";
export const L1_DUMMY_CONTRACT_ADDRESS = "0x75b33806dCdb0fC7BB06fd1121a3358a88EdC5E9";

export const SHOMEI_ENDPOINT = null;
export const SHOMEI_FRONTEND_ENDPOINT = null;
export const TRANSACTION_EXCLUSION_ENDPOINT = null;
