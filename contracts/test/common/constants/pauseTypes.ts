import {
  PAUSE_ALL_ROLE,
  UNPAUSE_ALL_ROLE,
  PAUSE_L1_L2_ROLE,
  UNPAUSE_L1_L2_ROLE,
  PAUSE_L2_L1_ROLE,
  UNPAUSE_L2_L1_ROLE,
  PAUSE_L2_BLOB_SUBMISSION_ROLE,
  UNPAUSE_L2_BLOB_SUBMISSION_ROLE,
  PAUSE_FINALIZE_WITHPROOF_ROLE,
  UNPAUSE_FINALIZE_WITHPROOF_ROLE,
  PAUSE_INITIATE_TOKEN_BRIDGING_ROLE,
  UNPAUSE_INITIATE_TOKEN_BRIDGING_ROLE,
  PAUSE_COMPLETE_TOKEN_BRIDGING_ROLE,
  UNPAUSE_COMPLETE_TOKEN_BRIDGING_ROLE,
} from "./roles";

export const GENERAL_PAUSE_TYPE = 1;
export const L1_L2_PAUSE_TYPE = 2;
export const L2_L1_PAUSE_TYPE = 3;
export const BLOB_SUBMISSION_PAUSE_TYPE = 4;
export const CALLDATA_SUBMISSION_PAUSE_TYPE = 5;
export const FINALIZATION_PAUSE_TYPE = 6;
export const INITIATE_TOKEN_BRIDGING_PAUSE_TYPE = 7;
export const COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE = 8;

export const pauseTypeRoles = [
  { pauseType: GENERAL_PAUSE_TYPE, role: PAUSE_ALL_ROLE },
  { pauseType: L1_L2_PAUSE_TYPE, role: PAUSE_L1_L2_ROLE },
  { pauseType: L2_L1_PAUSE_TYPE, role: PAUSE_L2_L1_ROLE },
  { pauseType: BLOB_SUBMISSION_PAUSE_TYPE, role: PAUSE_L2_BLOB_SUBMISSION_ROLE },
  { pauseType: CALLDATA_SUBMISSION_PAUSE_TYPE, role: PAUSE_L2_BLOB_SUBMISSION_ROLE },
  { pauseType: FINALIZATION_PAUSE_TYPE, role: PAUSE_FINALIZE_WITHPROOF_ROLE },
  { pauseType: INITIATE_TOKEN_BRIDGING_PAUSE_TYPE, role: PAUSE_INITIATE_TOKEN_BRIDGING_ROLE },
  { pauseType: COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE, role: PAUSE_COMPLETE_TOKEN_BRIDGING_ROLE },
];

export const unpauseTypeRoles = [
  { pauseType: GENERAL_PAUSE_TYPE, role: UNPAUSE_ALL_ROLE },
  { pauseType: L1_L2_PAUSE_TYPE, role: UNPAUSE_L1_L2_ROLE },
  { pauseType: L2_L1_PAUSE_TYPE, role: UNPAUSE_L2_L1_ROLE },
  { pauseType: BLOB_SUBMISSION_PAUSE_TYPE, role: UNPAUSE_L2_BLOB_SUBMISSION_ROLE },
  { pauseType: CALLDATA_SUBMISSION_PAUSE_TYPE, role: UNPAUSE_L2_BLOB_SUBMISSION_ROLE },
  { pauseType: FINALIZATION_PAUSE_TYPE, role: UNPAUSE_FINALIZE_WITHPROOF_ROLE },
  { pauseType: INITIATE_TOKEN_BRIDGING_PAUSE_TYPE, role: UNPAUSE_INITIATE_TOKEN_BRIDGING_ROLE },
  { pauseType: COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE, role: UNPAUSE_COMPLETE_TOKEN_BRIDGING_ROLE },
];