import { parseUnits } from "ethers";
import { ethers } from "hardhat";

export const MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS = 2000;
export const TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS = 2500;
export const MINIMUM_WITHDRAWAL_RESERVE_AMOUNT = ethers.parseEther("1000");
export const TARGET_WITHDRAWAL_RESERVE_AMOUNT = ethers.parseEther("1250");
export const MAX_BPS = 10000n;

// Values from constructor params for Lido PreDepositGuarantee.sol Hoodi deployment - https://hoodi.etherscan.io/address/0x8b289fc1af2bbc589f5990b94061d851c48683a3#code
export const GI_FIRST_VALIDATOR_PREV = "0x0000000000000000000000000000000000000000000000000096000000000028";
export const GI_FIRST_VALIDATOR_CURR = "0x0000000000000000000000000000000000000000000000000096000000000028";
export const PIVOT_SLOT = 0;

// YieldProviderVendor enum
// export const UNUSED_YIELD_PROVIDER_VENDOR = 0;
// export const LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR = 1;

export const enum YieldProviderVendor {
  UNUSED_YIELD_PROVIDER_VENDOR = 0,
  LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR = 1,
}

// ProgressOssificationResult enum
export const enum ProgressOssificationResult {
  REINITIATED = 0,
  NOOP = 1,
  COMPLETE = 2,
}

export const enum OperationType {
  FUND_YIELD_PROVIDER = 0,
  REPORT_YIELD = 1,
}

export const FAR_FUTURE_EXIT_EPOCH = 18446744073709551615n;
export const SHARD_COMMITTEE_PERIOD = 256n;
export const SLOTS_PER_EPOCH = 32n;

export const THIRTY_TWO_ETH_IN_GWEI = 32000000000n;
export const MAX_0X2_VALIDATOR_EFFECTIVE_BALANCE_GWEI = parseUnits("2048", "gwei");

export const VALIDATOR_WITNESS_TYPE =
  "tuple(bytes32[] proof, uint256 validatorIndex, uint64 effectiveBalance, uint64 childBlockTimestamp, uint64 slot, uint64 proposerIndex, uint64 activationEpoch, uint64 activationEligibilityEpoch)";
