import { ethers } from "ethers";
import * as dotenv from 'dotenv'
dotenv.config()

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
export const ACCOUNT_0_PRIVATE_KEY = process.env.LOCAL_ACCOUNT_0_PRIVATE_KEY?? "";
export const ACCOUNT_1 = "";

export const OPERATOR_0 = "0xd0584d4d37157f7105a4b41ed8ecbdfafdb2547f";
export const OPERATOR_0_PRIVATE_KEY = process.env.LOCAL_OPERATOR_0_PRIVATE_KEY?? "";


export const ZKEVMV2_CONTRACT_ADDRESS = "0xFE63fc3C8898F83B1c5F199133f89bDBA88B1C37";
export const ZKEVMV2_INITIAL_STATE_ROOT_HASH = "0x0000000000000000000000000000000000000000000000000000000000000000";
export const ZKEVMV2_INITIAL_L2_BLOCK_NR = "123456";
export const ZKEVMV2_SECURITY_COUNCIL = OPERATOR_0;
export const ZKEVMV2_OPERATORS = [OPERATOR_0];
export const ZKEVMV2_RATE_LIMIT_PERIOD = "86400"; //24Hours in seconds
export const ZKEVMV2_RATE_LIMIT_AMOUNT = "1000000000000000000000"; //1000ETH

export const TRANSACTION_DECODER_ADDRESS = "";
export const PLONK_VERIFIER_ADDRESS = "";
export const MESSAGE_SERVICE_ADDRESS = "0xa2d2C55e9B7054d9C4EA075Df35935BeA1693e27";
export const DUMMY_CONTRACT_ADDRESS = "";
