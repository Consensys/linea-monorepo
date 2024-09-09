import { ethers } from "ethers";
import * as dotenv from "dotenv";
dotenv.config();

export const TRANSACTION_CALLDATA_LIMIT = 30000;
export const L1_RPC_URL = "https://l1-rpc.dev.zkevm.consensys.net/";
export const L2_RPC_URL = "https://archive.dev.zkevm.consensys.net/";
export const CHAIN_ID = 59139;

export function getL1Provider() {
  return new ethers.providers.JsonRpcProvider(L1_RPC_URL);
}

export function getL2Provider() {
  return new ethers.providers.JsonRpcProvider(L2_RPC_URL);
}

export const DEPLOYER_ACCOUNT_PRIVATE_KEY = process.env.LOCAL_DEPLOYER_ACCOUNT_PRIVATE_KEY ?? "";

export const INITIAL_WITHDRAW_LIMIT = ethers.utils.parseEther("5");

export const ACCOUNT_0 = "0xc8c92fe825d8930b9357c006e0af160dfa727a62";
export const ACCOUNT_0_PRIVATE_KEY = process.env.LOCAL_ACCOUNT_0_PRIVATE_KEY ?? "";
export const ACCOUNT_1 = "";

export const OPERATOR_0 = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8";
export const OPERATOR_0_PRIVATE_KEY = process.env.LOCAL_OPERATOR_0_PRIVATE_KEY ?? "";

export const LINEA_ROLLUP_CONTRACT_ADDRESS = "0xFE63fc3C8898F83B1c5F199133f89bDBA88B1C37";
export const LINEA_ROLLUP_INITIAL_STATE_ROOT_HASH =
  "0x0000000000000000000000000000000000000000000000000000000000000000";
export const LINEA_ROLLUP_INITIAL_L2_BLOCK_NR = "123456";
export const LINEA_ROLLUP_SECURITY_COUNCIL = OPERATOR_0;
export const LINEA_ROLLUP_OPERATORS = [OPERATOR_0];
export const LINEA_ROLLUP_RATE_LIMIT_PERIOD = "86400"; //24Hours in seconds
export const LINEA_ROLLUP_RATE_LIMIT_AMOUNT = "1000000000000000000000"; //1000ETH

export const TRANSACTION_DECODER_ADDRESS = "";
export const PLONK_VERIFIER_ADDRESS = "";
export const MESSAGE_SERVICE_ADDRESS = "0xa2d2C55e9B7054d9C4EA075Df35935BeA1693e27";
export const DUMMY_CONTRACT_ADDRESS = "";

export const SHOMEI_ENDPOINT = null;
export const SHOMEI_FRONTEND_ENDPOINT = null;
export const TRANSACTION_EXCLUSION_ENDPOINT = null;
