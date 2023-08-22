import { ethers } from "hardhat";

// TODO FIX IMPORT
const generateKeccak256Hash = (str: string) => ethers.utils.keccak256(ethers.utils.toUtf8Bytes(str));

// TimeLock roles
export const TIMELOCK_ADMIN_ROLE = generateKeccak256Hash("TIMELOCK_ADMIN_ROLE");
export const PROPOSER_ROLE = generateKeccak256Hash("PROPOSER_ROLE");
export const EXECUTOR_ROLE = generateKeccak256Hash("EXECUTOR_ROLE");
export const CANCELLER_ROLE = generateKeccak256Hash("CANCELLER_ROLE");

// Roles hashes
export const DEFAULT_ADMIN_ROLE = ethers.constants.HashZero;
export const RATE_LIMIT_SETTER_ROLE = generateKeccak256Hash("RATE_LIMIT_SETTER_ROLE");
export const L1_L2_MESSAGE_SETTER_ROLE = generateKeccak256Hash("L1_L2_MESSAGE_SETTER_ROLE");
export const PAUSE_MANAGER_ROLE = generateKeccak256Hash("PAUSE_MANAGER_ROLE");
export const MINIMUM_FEE_SETTER_ROLE = generateKeccak256Hash("MINIMUM_FEE_SETTER_ROLE");
export const OPERATOR_ROLE = generateKeccak256Hash("OPERATOR_ROLE");
export const BAD_STARTING_HASH = generateKeccak256Hash("BAD_STARTING_HASH");

export const INITIAL_MIGRATION_BLOCK = 0;
export const L1_L2_PAUSE_TYPE = generateKeccak256Hash("L1_L2_PAUSE_TYPE");
export const L2_L1_PAUSE_TYPE = generateKeccak256Hash("L2_L1_PAUSE_TYPE");
export const PROVING_SYSTEM_PAUSE_TYPE = generateKeccak256Hash("PROVING_SYSTEM_PAUSE_TYPE");
export const GENERAL_PAUSE_TYPE = generateKeccak256Hash("GENERAL_PAUSE_TYPE");

export const INBOX_STATUS_UNKNOWN = 0;
export const INBOX_STATUS_RECEIVED = 1;
export const INBOX_STATUS_CLAIMED = 2;

export const OUTBOX_STATUS_UNKNOWN = 0;
export const OUTBOX_STATUS_SENT = 1;
export const OUTBOX_STATUS_RECEIVED = 2;

export const ONE_DAY_IN_SECONDS = 86_400;
export const INITIAL_WITHDRAW_LIMIT = ethers.utils.parseEther("5");

export const MESSAGE_VALUE_1ETH = ethers.utils.parseEther("1");
export const ZERO_VALUE = 0;
export const MESSAGE_FEE = ethers.utils.parseEther("0.05");
export const LOW_NO_REFUND_MESSAGE_FEE = ethers.utils.parseEther("0.00001");
export const MINIMUM_FEE = ethers.utils.parseEther("0.1");
export const DEFAULT_MESSAGE_NONCE = ethers.utils.parseEther("123456789");
export const SAMPLE_FUNCTION_CALLDATA = generateKeccak256Hash("callThisFunction()").substring(0, 10); //0x + 4bytes
export const EMPTY_CALLDATA = "0x";
export const BLOCK_COINBASE = "0xc014ba5ec014ba5ec014ba5ec014ba5ec014ba5e";

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
