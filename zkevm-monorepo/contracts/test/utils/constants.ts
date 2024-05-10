import { ethers } from "hardhat";
import { generateKeccak256 } from "./helpers";

export const HASH_ZERO = ethers.ZeroHash;
export const ADDRESS_ZERO = ethers.ZeroAddress;
export const HASH_WITHOUT_ZERO_FIRST_BYTE = "0xf887bbc07b0e849fb625aafadf4cb6b65b98e492fbb689705312bf1db98ead7f";

// Linea XP Token roles
export const MINTER_ROLE = generateKeccak256(["string"], ["MINTER_ROLE"], true);

// TimeLock roles
export const TIMELOCK_ADMIN_ROLE = generateKeccak256(["string"], ["TIMELOCK_ADMIN_ROLE"], true);
export const PROPOSER_ROLE = generateKeccak256(["string"], ["PROPOSER_ROLE"], true);
export const EXECUTOR_ROLE = generateKeccak256(["string"], ["EXECUTOR_ROLE"], true);
export const CANCELLER_ROLE = generateKeccak256(["string"], ["CANCELLER_ROLE"], true);

// Roles hashes
export const DEFAULT_ADMIN_ROLE = HASH_ZERO;
export const RATE_LIMIT_SETTER_ROLE = generateKeccak256(["string"], ["RATE_LIMIT_SETTER_ROLE"], true);
export const L1_L2_MESSAGE_SETTER_ROLE = generateKeccak256(["string"], ["L1_L2_MESSAGE_SETTER_ROLE"], true);
export const PAUSE_MANAGER_ROLE = generateKeccak256(["string"], ["PAUSE_MANAGER_ROLE"], true);
export const MINIMUM_FEE_SETTER_ROLE = generateKeccak256(["string"], ["MINIMUM_FEE_SETTER_ROLE"], true);
export const OPERATOR_ROLE = generateKeccak256(["string"], ["OPERATOR_ROLE"], true);
export const VERIFIER_SETTER_ROLE = generateKeccak256(["string"], ["VERIFIER_SETTER_ROLE"], true);
export const L1_MERKLE_ROOTS_SETTER_ROLE = generateKeccak256(["string"], ["L1_MERKLE_ROOTS_SETTER_ROLE"], true);
export const L2_MERKLE_ROOTS_SETTER_ROLE = generateKeccak256(["string"], ["L2_MERKLE_ROOTS_SETTER_ROLE"], true);
export const BAD_STARTING_HASH = generateKeccak256(["string"], ["BAD_STARTING_HASH"], true);

export const GENERAL_PAUSE_TYPE = 1;
export const L1_L2_PAUSE_TYPE = 2;
export const L2_L1_PAUSE_TYPE = 3;
export const PROVING_SYSTEM_PAUSE_TYPE = 4;

// Message statuses
export const INBOX_STATUS_UNKNOWN = 0;
export const INBOX_STATUS_RECEIVED = 1;
export const INBOX_STATUS_CLAIMED = 2;

export const OUTBOX_STATUS_UNKNOWN = 0;
export const OUTBOX_STATUS_SENT = 1;
export const OUTBOX_STATUS_RECEIVED = 2;

export const INITIAL_MIGRATION_BLOCK = 0;
export const ONE_DAY_IN_SECONDS = 86_400;
export const INITIAL_WITHDRAW_LIMIT = ethers.parseEther("5");
export const GENESIS_L2_TIMESTAMP = 0;
export const TEST_PUBLIC_VERIFIER_INDEX = 0;

export const MESSAGE_VALUE_1ETH = ethers.parseEther("1");
export const ZERO_VALUE = 0;
export const MESSAGE_FEE = ethers.parseEther("0.05");
export const LOW_NO_REFUND_MESSAGE_FEE = ethers.parseEther("0.00001");
export const MINIMUM_FEE = ethers.parseEther("0.1");
export const DEFAULT_MESSAGE_NONCE = ethers.parseEther("123456789");
export const SAMPLE_FUNCTION_CALLDATA = generateKeccak256(["string"], ["callThisFunction()"], true).substring(0, 10); //0x + 4bytes
export const EMPTY_CALLDATA = "0x";
export const BLOCK_COINBASE = "0xc014ba5ec014ba5ec014ba5ec014ba5ec014ba5e";

export const DEFAULT_SUBMISSION_DATA = {
  dataParentHash: HASH_ZERO,
  compressedData: "0x",
  finalBlockInData: 0n,
  firstBlockInData: 0n,
  parentStateRootHash: HASH_ZERO,
  finalStateRootHash: HASH_ZERO,
  snarkHash: HASH_ZERO,
};

// TODO CLEANUP TO MAKE THIS DYNAMIC AND NOT CONSTANT
export const Add_L1L2_Message_Hashes_Calldata_With_Empty_Array =
  "0xf4b476e10000000000000000000000000000000000000000000000000000000000000000";
export const Add_L1L2_Message_Hashes_Calldata_With_One_Hash =
  "0xf4b476e100000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001f57d2ce8b7fc0df7ae7cbddaa706242a118bd8b649abccfecfb3f8e419729ca9";
export const Single_Item_L1L2_HashArray = ["0xf57d2ce8b7fc0df7ae7cbddaa706242a118bd8b649abccfecfb3f8e419729ca9"];

// TODO CLEANUP TO MAKE THIS DYNAMIC AND NOT CONSTANT
export const Add_L1L2_Message_Hashes_Calldata_With_Five_Hashes =
  "0xf4b476e100000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000005f887bbc07b0e849fb625aafadf4cb6b65b98e492fbb689705312bf1db98ead7fdd87bbc07b0e849fb625aafadf4cb6b65b98e492fbb689705312bf1db98ead7faa87bbc07b0e849fb625aafadf4cb6b65b98e492fbb689705312bf1db98ead7fcc87bbc07b0e849fb625aafadf4cb6b65b98e492fbb689705312bf1db98ead7f1187bbc07b0e849fb625aafadf4cb6b65b98e492fbb689705312bf1db98ead7f";
export const L1L2_FiveHashes = [
  "0xf887bbc07b0e849fb625aafadf4cb6b65b98e492fbb689705312bf1db98ead7f",
  "0xdd87bbc07b0e849fb625aafadf4cb6b65b98e492fbb689705312bf1db98ead7f",
  "0xaa87bbc07b0e849fb625aafadf4cb6b65b98e492fbb689705312bf1db98ead7f",
  "0xcc87bbc07b0e849fb625aafadf4cb6b65b98e492fbb689705312bf1db98ead7f",
  "0x1187bbc07b0e849fb625aafadf4cb6b65b98e492fbb689705312bf1db98ead7f",
];
