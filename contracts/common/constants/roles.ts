import { HASH_ZERO } from "./general";
import { generateKeccak256 } from "../helpers";

// Roles hashes
export const DEFAULT_ADMIN_ROLE = HASH_ZERO;
export const FUNCTION_EXECUTOR_ROLE = generateKeccak256(["string"], ["FUNCTION_EXECUTOR_ROLE"], { encodePacked: true });
export const RATE_LIMIT_SETTER_ROLE = generateKeccak256(["string"], ["RATE_LIMIT_SETTER_ROLE"], { encodePacked: true });
export const USED_RATE_LIMIT_RESETTER_ROLE = generateKeccak256(["string"], ["USED_RATE_LIMIT_RESETTER_ROLE"], {
  encodePacked: true,
});
export const L1_L2_MESSAGE_SETTER_ROLE = generateKeccak256(["string"], ["L1_L2_MESSAGE_SETTER_ROLE"], {
  encodePacked: true,
});
export const PAUSE_ALL_ROLE = generateKeccak256(["string"], ["PAUSE_ALL_ROLE"], { encodePacked: true });
export const UNPAUSE_ALL_ROLE = generateKeccak256(["string"], ["UNPAUSE_ALL_ROLE"], { encodePacked: true });
export const PAUSE_L1_L2_ROLE = generateKeccak256(["string"], ["PAUSE_L1_L2_ROLE"], { encodePacked: true });
export const UNPAUSE_L1_L2_ROLE = generateKeccak256(["string"], ["UNPAUSE_L1_L2_ROLE"], { encodePacked: true });
export const PAUSE_L2_L1_ROLE = generateKeccak256(["string"], ["PAUSE_L2_L1_ROLE"], { encodePacked: true });
export const UNPAUSE_L2_L1_ROLE = generateKeccak256(["string"], ["UNPAUSE_L2_L1_ROLE"], { encodePacked: true });
export const PAUSE_BLOB_SUBMISSION_ROLE = generateKeccak256(["string"], ["PAUSE_BLOB_SUBMISSION_ROLE"], {
  encodePacked: true,
});
export const UNPAUSE_BLOB_SUBMISSION_ROLE = generateKeccak256(["string"], ["UNPAUSE_BLOB_SUBMISSION_ROLE"], {
  encodePacked: true,
});
export const PAUSE_FINALIZE_WITHPROOF_ROLE = generateKeccak256(["string"], ["PAUSE_FINALIZE_WITHPROOF_ROLE"], {
  encodePacked: true,
});
export const UNPAUSE_FINALIZE_WITHPROOF_ROLE = generateKeccak256(["string"], ["UNPAUSE_FINALIZE_WITHPROOF_ROLE"], {
  encodePacked: true,
});
export const MINIMUM_FEE_SETTER_ROLE = generateKeccak256(["string"], ["MINIMUM_FEE_SETTER_ROLE"], {
  encodePacked: true,
});
export const OPERATOR_ROLE = generateKeccak256(["string"], ["OPERATOR_ROLE"], { encodePacked: true });
export const VERIFIER_SETTER_ROLE = generateKeccak256(["string"], ["VERIFIER_SETTER_ROLE"], { encodePacked: true });
export const VERIFIER_UNSETTER_ROLE = generateKeccak256(["string"], ["VERIFIER_UNSETTER_ROLE"], { encodePacked: true });
export const L1_MERKLE_ROOTS_SETTER_ROLE = generateKeccak256(["string"], ["L1_MERKLE_ROOTS_SETTER_ROLE"], {
  encodePacked: true,
});
export const L2_MERKLE_ROOTS_SETTER_ROLE = generateKeccak256(["string"], ["L2_MERKLE_ROOTS_SETTER_ROLE"], {
  encodePacked: true,
});
export const PAUSE_INITIATE_TOKEN_BRIDGING_ROLE = generateKeccak256(
  ["string"],
  ["PAUSE_INITIATE_TOKEN_BRIDGING_ROLE"],
  { encodePacked: true },
);
export const PAUSE_COMPLETE_TOKEN_BRIDGING_ROLE = generateKeccak256(
  ["string"],
  ["PAUSE_COMPLETE_TOKEN_BRIDGING_ROLE"],
  { encodePacked: true },
);
export const UNPAUSE_INITIATE_TOKEN_BRIDGING_ROLE = generateKeccak256(
  ["string"],
  ["UNPAUSE_INITIATE_TOKEN_BRIDGING_ROLE"],
  { encodePacked: true },
);
export const UNPAUSE_COMPLETE_TOKEN_BRIDGING_ROLE = generateKeccak256(
  ["string"],
  ["UNPAUSE_COMPLETE_TOKEN_BRIDGING_ROLE"],
  { encodePacked: true },
);
export const SET_REMOTE_TOKENBRIDGE_ROLE = generateKeccak256(["string"], ["SET_REMOTE_TOKENBRIDGE_ROLE"], {
  encodePacked: true,
});
export const SET_RESERVED_TOKEN_ROLE = generateKeccak256(["string"], ["SET_RESERVED_TOKEN_ROLE"], {
  encodePacked: true,
});
export const REMOVE_RESERVED_TOKEN_ROLE = generateKeccak256(["string"], ["REMOVE_RESERVED_TOKEN_ROLE"], {
  encodePacked: true,
});
export const SET_CUSTOM_CONTRACT_ROLE = generateKeccak256(["string"], ["SET_CUSTOM_CONTRACT_ROLE"], {
  encodePacked: true,
});
export const SET_MESSAGE_SERVICE_ROLE = generateKeccak256(["string"], ["SET_MESSAGE_SERVICE_ROLE"], {
  encodePacked: true,
});

export const BASE_ROLES = [PAUSE_ALL_ROLE, UNPAUSE_ALL_ROLE];

export const LINEA_ROLLUP_ROLES = [
  ...BASE_ROLES,
  VERIFIER_SETTER_ROLE,
  VERIFIER_UNSETTER_ROLE,
  RATE_LIMIT_SETTER_ROLE,
  USED_RATE_LIMIT_RESETTER_ROLE,
  PAUSE_L1_L2_ROLE,
  PAUSE_L2_L1_ROLE,
  UNPAUSE_L1_L2_ROLE,
  UNPAUSE_L2_L1_ROLE,
  PAUSE_BLOB_SUBMISSION_ROLE,
  UNPAUSE_BLOB_SUBMISSION_ROLE,
  PAUSE_FINALIZE_WITHPROOF_ROLE,
  UNPAUSE_FINALIZE_WITHPROOF_ROLE,
];

export const L2_MESSAGE_SERVICE_ROLES = [
  ...BASE_ROLES,
  MINIMUM_FEE_SETTER_ROLE,
  RATE_LIMIT_SETTER_ROLE,
  USED_RATE_LIMIT_RESETTER_ROLE,
  PAUSE_L1_L2_ROLE,
  PAUSE_L2_L1_ROLE,
  UNPAUSE_L1_L2_ROLE,
  UNPAUSE_L2_L1_ROLE,
  L1_L2_MESSAGE_SETTER_ROLE,
];

export const TOKEN_BRIDGE_ROLES = [
  ...BASE_ROLES,
  SET_MESSAGE_SERVICE_ROLE,
  SET_REMOTE_TOKENBRIDGE_ROLE,
  SET_RESERVED_TOKEN_ROLE,
  REMOVE_RESERVED_TOKEN_ROLE,
  SET_CUSTOM_CONTRACT_ROLE,
  PAUSE_INITIATE_TOKEN_BRIDGING_ROLE,
  UNPAUSE_INITIATE_TOKEN_BRIDGING_ROLE,
  PAUSE_COMPLETE_TOKEN_BRIDGING_ROLE,
  UNPAUSE_COMPLETE_TOKEN_BRIDGING_ROLE,
];
