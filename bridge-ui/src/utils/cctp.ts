import { getAddress } from "viem";

export const CCTP_TOKEN_MESSENGER = getAddress("0x8FE6B999Dc680CcFDD5Bf7EB0974218be2542DAA");
export const CCTP_TRANSFER_MAX_FEE = 500n;
export const CCTP_MIN_FINALITY_THRESHOLD = 1000; // 1000 Fast transfer, 2000 Standard transfer
