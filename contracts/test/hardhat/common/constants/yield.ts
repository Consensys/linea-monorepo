import { ethers } from "hardhat";

export const MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS = 2000;
export const TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS = 2500;
export const MINIMUM_WITHDRAWAL_RESERVE_AMOUNT = ethers.parseEther("1000");
export const TARGET_WITHDRAWAL_RESERVE_AMOUNT = ethers.parseEther("1250");
export const MAX_BPS = 10000n;

// Values from constructor params for Lido PreDepositGuarantee.sol Hoodi deployment - https://hoodi.etherscan.io/address/0x8b289fc1af2bbc589f5990b94061d851c48683a3#code
export const GI_FIRST_VALIDATOR = "0x0000000000000000000000000000000000000000000000000096000000000028";
export const GI_FIRST_VALIDATOR_AFTER_CHANGE = "0x0000000000000000000000000000000000000000000000000096000000000028";
export const CHANGE_SLOT = 0;

// YieldProviderVendor enum
export const UNUSED_YIELD_PROVIDER_VENDOR = 0;
export const LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR = 1;

// YieldProviderRegistrationError enum
export const LIDO_DASHBOARD_NOT_LINKED_TO_VAULT = 0;
export const LIDO_VAULT_IS_EXPECTED_RECEIVE_CALLER_AND_OSSIFIED_ENTRYPOINT = 1;

// OperationType enum
export const REPORT_YIELD_OPERATION_TYPE = 0;
