import { ethers } from "ethers";
import { generateKeccak256 } from "./utils";

export const ROLLING_HASH_UPDATED_EVENT_SIGNATURE =
  "0xea3b023b4c8680d4b4824f0143132c95476359a2bb70a81d6c5a36f6918f6339";
export const OPERATOR_ROLE = "0x97667070c54ef182b0f5858b034beac1b6f3089aa2d3188bb1e8929f4fa9b929";
export const VERIFIER_SETTER_ROLE = "0x32937fd5162e282df7e9a14a5073a2425321c7966eaf70ed6c838a1006d84c4c";
export const MESSAGE_SENT_EVENT_SIGNATURE = "0xe856c2b8bd4eb0027ce32eeaf595c21b0b6b4644b326e5b7bd80a1cf8db72e6c";

export const HASH_ZERO = ethers.ZeroHash;
export const EMPTY_CONTRACT_CODE = "0x";
export const TRANSACTION_CALLDATA_LIMIT = 30_000;

export const PAUSE_ALL_ROLE = generateKeccak256(["string"], ["PAUSE_ALL_ROLE"], true);
export const UNPAUSE_ALL_ROLE = generateKeccak256(["string"], ["UNPAUSE_ALL_ROLE"], true);
export const PAUSE_L1_L2_ROLE = generateKeccak256(["string"], ["PAUSE_L1_L2_ROLE"], true);
export const UNPAUSE_L1_L2_ROLE = generateKeccak256(["string"], ["UNPAUSE_L1_L2_ROLE"], true);
export const PAUSE_L2_L1_ROLE = generateKeccak256(["string"], ["PAUSE_L2_L1_ROLE"], true);
export const UNPAUSE_L2_L1_ROLE = generateKeccak256(["string"], ["UNPAUSE_L2_L1_ROLE"], true);
export const PAUSE_BLOB_SUBMISSION_ROLE = generateKeccak256(["string"], ["PAUSE_BLOB_SUBMISSION_ROLE"], true);
export const UNPAUSE_BLOB_SUBMISSION_ROLE = generateKeccak256(["string"], ["UNPAUSE_BLOB_SUBMISSION_ROLE"], true);
export const PAUSE_FINALIZATION_ROLE = generateKeccak256(["string"], ["PAUSE_FINALIZATION_ROLE"], true);
export const UNPAUSE_FINALIZATION_ROLE = generateKeccak256(["string"], ["UNPAUSE_FINALIZATION_ROLE"], true);
export const USED_RATE_LIMIT_RESETTER_ROLE = generateKeccak256(["string"], ["USED_RATE_LIMIT_RESETTER_ROLE"], true);
export const VERIFIER_UNSETTER_ROLE = generateKeccak256(["string"], ["VERIFIER_UNSETTER_ROLE"], true);

export const LINEA_ROLLUP_V6_ROLES = [
  PAUSE_ALL_ROLE,
  PAUSE_L1_L2_ROLE,
  PAUSE_L2_L1_ROLE,
  UNPAUSE_ALL_ROLE,
  UNPAUSE_L1_L2_ROLE,
  UNPAUSE_L2_L1_ROLE,
  PAUSE_BLOB_SUBMISSION_ROLE,
  UNPAUSE_BLOB_SUBMISSION_ROLE,
  PAUSE_FINALIZATION_ROLE,
  UNPAUSE_FINALIZATION_ROLE,
  USED_RATE_LIMIT_RESETTER_ROLE,
  VERIFIER_UNSETTER_ROLE,
];

export const GENERAL_PAUSE_TYPE = 1;
export const L1_L2_PAUSE_TYPE = 2;
export const L2_L1_PAUSE_TYPE = 3;
export const BLOB_SUBMISSION_PAUSE_TYPE = 4;
export const CALLDATA_SUBMISSION_PAUSE_TYPE = 5;
export const FINALIZATION_PAUSE_TYPE = 6;

export const BASE_PAUSE_TYPES_ROLES = [{ pauseType: GENERAL_PAUSE_TYPE, role: PAUSE_ALL_ROLE }];
export const BASE_UNPAUSE_TYPES_ROLES = [{ pauseType: GENERAL_PAUSE_TYPE, role: UNPAUSE_ALL_ROLE }];

export const LINEA_ROLLUP_PAUSE_TYPES_ROLES = [
  ...BASE_PAUSE_TYPES_ROLES,
  { pauseType: L1_L2_PAUSE_TYPE, role: PAUSE_L1_L2_ROLE },
  { pauseType: L2_L1_PAUSE_TYPE, role: PAUSE_L2_L1_ROLE },
  { pauseType: BLOB_SUBMISSION_PAUSE_TYPE, role: PAUSE_BLOB_SUBMISSION_ROLE },
  { pauseType: CALLDATA_SUBMISSION_PAUSE_TYPE, role: PAUSE_BLOB_SUBMISSION_ROLE },
  { pauseType: FINALIZATION_PAUSE_TYPE, role: PAUSE_FINALIZATION_ROLE },
];

export const LINEA_ROLLUP_UNPAUSE_TYPES_ROLES = [
  ...BASE_UNPAUSE_TYPES_ROLES,
  { pauseType: L1_L2_PAUSE_TYPE, role: UNPAUSE_L1_L2_ROLE },
  { pauseType: L2_L1_PAUSE_TYPE, role: UNPAUSE_L2_L1_ROLE },
  { pauseType: BLOB_SUBMISSION_PAUSE_TYPE, role: UNPAUSE_BLOB_SUBMISSION_ROLE },
  { pauseType: CALLDATA_SUBMISSION_PAUSE_TYPE, role: UNPAUSE_BLOB_SUBMISSION_ROLE },
  { pauseType: FINALIZATION_PAUSE_TYPE, role: UNPAUSE_FINALIZATION_ROLE },
];
