
import { ethers } from "ethers";
import * as dotenv from 'dotenv'
dotenv.config()

export const TRANSACTION_CALLDATA_LIMIT = 30000;
export const L1_RPC_URL = "http://localhost:8445";
export const L2_RPC_URL = "http://localhost:8545";
export const CHAIN_ID = 1337;


export function getL1Provider() {
    return new ethers.providers.JsonRpcProvider(L1_RPC_URL);
}

export function getL2Provider() {
    return new ethers.providers.JsonRpcProvider(L2_RPC_URL);
}

// WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
export const DEPLOYER_ACCOUNT_PRIVATE_KEY = "0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae";

export const INITIAL_WITHDRAW_LIMIT = ethers.utils.parseEther("5");

export const ACCOUNT_0 = "0x1b9abeec3215d8ade8a33607f2cf0f4f60e5f0d0";
// WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
export const ACCOUNT_0_PRIVATE_KEY = "0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae";

export const OPERATOR_0 = "0xd0584d4d37157f7105a4b41ed8ecbdfafdb2547f";

// WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
export const OPERATOR_0_PRIVATE_KEY = "0x202454d1b4e72c41ebf58150030f649648d3cf5590297fb6718e27039ed9c86d";

export const ZKEVMV2_INITIAL_STATE_ROOT_HASH = "0x0000000000000000000000000000000000000000000000000000000000000000";
export const ZKEVMV2_INITIAL_L2_BLOCK_NR = "123456";
export const ZKEVMV2_SECURITY_COUNCIL = OPERATOR_0;
export const ZKEVMV2_OPERATORS = [OPERATOR_0];
export const ZKEVMV2_RATE_LIMIT_PERIOD = "86400"; //24Hours in seconds
export const ZKEVMV2_RATE_LIMIT_AMOUNT = "1000000000000000000000"; //1000ETH
export const ZKEVMV2_CONTRACT_ADDRESS = "0xC737F2334651ea85A72D8DA9d933c821A8377F9f";
export const MESSAGE_SERVICE_ADDRESS = "0xe537D669CA013d86EBeF1D64e40fC74CADC91987";
