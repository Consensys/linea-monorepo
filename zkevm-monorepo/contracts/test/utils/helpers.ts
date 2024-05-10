import { time as networkTime } from "@nomicfoundation/hardhat-network-helpers";
import fs from "fs";
import { ethers } from "hardhat";
import path from "path";
import { DEFAULT_MESSAGE_NONCE, MESSAGE_FEE, MESSAGE_VALUE_1ETH } from "./constants";
import {
  DebugData,
  FormattedBlockData,
  RawBlockData,
  FinalizationData,
  SubmissionData,
  SubmissionAndCompressedData,
  CalldataSubmissionData,
} from "./types";

import firstCompressedDataContent from "../testData/compressedData/blocks-1-46.json";
import fourthCompressedDataContent from "../testData/compressedData/blocks-115-155.json";
import secondCompressedDataContent from "../testData/compressedData/blocks-47-81.json";
import thirdCompressedDataContent from "../testData/compressedData/blocks-82-114.json";

const COMPRESSED_SUBMISSION_DATA = [
  firstCompressedDataContent,
  secondCompressedDataContent,
  thirdCompressedDataContent,
  fourthCompressedDataContent,
];

import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { AbiCoder } from "ethers";
import { ILineaRollup } from "../../typechain-types";
import firstCompressedDataContent_multiple from "../testData/compressedData/multipleProofs/blocks-1-46.json";
import fourthCompressedDataContent_multiple from "../testData/compressedData/multipleProofs/blocks-120-153.json";
import secondCompressedDataContent_multiple from "../testData/compressedData/multipleProofs/blocks-47-81.json";
import thirdCompressedDataContent_multiple from "../testData/compressedData/multipleProofs/blocks-82-119.json";

const COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF = [
  firstCompressedDataContent_multiple,
  secondCompressedDataContent_multiple,
  thirdCompressedDataContent_multiple,
  fourthCompressedDataContent_multiple,
];

// TODO: to be removed. Use the generateKeccak256 function instead
export const generateKeccak256Hash = (str: string) => ethers.keccak256(ethers.toUtf8Bytes(str));

export const generateKeccak256 = (types: string[], values: unknown[], packed?: boolean) =>
  ethers.keccak256(encodeData(types, values, packed));

export const encodeData = (types: string[], values: unknown[], packed?: boolean) => {
  if (packed) {
    return ethers.solidityPacked(types, values);
  }
  return AbiCoder.defaultAbiCoder().encode(types, values);
};

export const generateNKeccak256Hashes = (str: string, numberOfHashToGenerate: number): string[] => {
  let arr: string[] = [];
  for (let i = 1; i < numberOfHashToGenerate + 1; i++) {
    arr = [...arr, generateKeccak256(["string"], [`${str}${i}`], true)];
  }
  return arr;
};

export const generateRandomBytes = (length: number): string => ethers.hexlify(ethers.randomBytes(length));

export async function encodeSendMessageLog(
  sender: SignerWithAddress,
  receiver: SignerWithAddress,
  messageHash: string,
  calldata: string,
) {
  const topic = ethers.id("MessageSent(address,address,uint256,uint256,uint256,bytes,bytes32)");
  const data = encodeData(
    ["address", "address", "uint256", "uint256", "uint256", "bytes", "bytes32"],
    [sender.address, receiver.address, MESSAGE_FEE, MESSAGE_VALUE_1ETH, DEFAULT_MESSAGE_NONCE, calldata, messageHash],
  );

  return {
    topic,
    data,
  };
}

export async function encodeSendMessage(
  sender: string,
  receiver: string,
  fee: bigint,
  amount: bigint,
  salt: bigint,
  calldata: string,
) {
  return encodeData(
    ["address", "address", "uint256", "uint256", "uint256", "bytes"],
    [sender, receiver, fee, amount, salt, calldata],
  );
}

export function calculateRollingHash(existingRollingHash: string, messageHash: string) {
  return generateKeccak256(["bytes32", "bytes32"], [existingRollingHash, messageHash]);
}

export function calculateRollingHashFromCollection(existingRollingHash: string, messageHashes: string[]) {
  return messageHashes.reduce((rollingHash, hash) => calculateRollingHash(rollingHash, hash), existingRollingHash);
}

export const range = (start: number, end: number) => Array.from(Array(end - start + 1).keys()).map((x) => x + start);

export function getProverTestData(
  folder: string,
  filename: string,
): {
  blocks: FormattedBlockData[];
  proverMode: string;
  parentStateRootHash: string;
  version: string;
  firstBlockNumber: number;
  proof: string;
  debugData: DebugData;
} {
  const testFilePath = path.resolve(__dirname, "..", "testData", folder, filename);
  const testData = JSON.parse(fs.readFileSync(testFilePath, "utf8"));

  return {
    blocks: testData.blocksData.map((block: RawBlockData) => ({
      blockRootHash: block.rootHash,
      transactions: block.rlpEncodedTransactions,
      l2BlockTimestamp: block.timestamp,
      l2ToL1MsgHashes: block.l2ToL1MsgHashes,
      fromAddresses: block.fromAddresses,
      batchReceptionIndices: block.batchReceptionIndices,
    })),
    proverMode: testData.proverMode,
    parentStateRootHash: testData.parentStateRootHash,
    version: testData.version,
    firstBlockNumber: testData.firstBlockNumber,
    proof: testData.proof,
    debugData: testData.DebugData,
  };
}

export function getRLPEncodeTransactions(filename: string): {
  shortEip1559Transaction: string;
  eip1559Transaction: string;
  eip1559TransactionHashes: string[];
  legacyTransaction: string;
  legacyTransactionHashes: string[];
  eip2930Transaction: string;
  eip2930TransactionHashes: string[];
} {
  const testFilePath = path.resolve(__dirname, "..", "testData", filename);
  const testData = JSON.parse(fs.readFileSync(testFilePath, "utf8"));

  return {
    shortEip1559Transaction: testData.shortEip1559Transaction,
    eip1559Transaction: testData.eip1559Transaction,
    eip1559TransactionHashes: testData.eip1559TransactionHashes,
    legacyTransaction: testData.legacyTransaction,
    legacyTransactionHashes: testData.legacyTransactionHashes,
    eip2930Transaction: testData.eip2930Transaction,
    eip2930TransactionHashes: testData.eip2930TransactionHashes,
  };
}

export function getTransactionsToBeDecoded(blocks: FormattedBlockData[]): string[] {
  const txsToBeDecoded = [];
  for (let i = 0; i < blocks.length; i++) {
    for (let j = 0; j < blocks[i].batchReceptionIndices.length; j++) {
      const txIndex = blocks[i].batchReceptionIndices[j];
      txsToBeDecoded.push(blocks[i].transactions[txIndex]);
    }
  }
  return txsToBeDecoded;
}

export const generateL2MessagingBlocksOffsets = (start: number, end: number) =>
  `0x${range(start, end)
    .map((num) => ethers.solidityPacked(["uint16"], [num]).slice(2))
    .join("")}`;

export async function generateFinalizationData(overrides?: Partial<FinalizationData>): Promise<FinalizationData> {
  return {
    aggregatedProof: generateRandomBytes(928),
    parentStateRootHash: generateRandomBytes(32),
    dataHashes: [generateRandomBytes(32)],
    dataParentHash: generateRandomBytes(32),
    finalBlockNumber: 812502n,
    lastFinalizedTimestamp: BigInt((await networkTime.latest()) - 2),
    finalTimestamp: BigInt(await networkTime.latest()),
    l1RollingHash: generateRandomBytes(32),
    l1RollingHashMessageNumber: 10n,
    l2MerkleRoots: [generateRandomBytes(32)],
    l2MerkleTreesDepth: 5n,
    l2MessagingBlocksOffsets: generateL2MessagingBlocksOffsets(1, 1),
    ...overrides,
  };
}

export function createLineaRollupFinalizationData(
  finalizationData: FinalizationData,
): ILineaRollup.FinalizationDataStruct {
  return {
    parentStateRootHash: finalizationData.parentStateRootHash,
    dataHashes: finalizationData.dataHashes,
    dataParentHash: finalizationData.dataParentHash,
    finalBlockNumber: finalizationData.finalBlockNumber,
    lastFinalizedTimestamp: finalizationData.lastFinalizedTimestamp,
    finalTimestamp: finalizationData.finalTimestamp,
    l1RollingHash: finalizationData.l1RollingHash,
    l1RollingHashMessageNumber: finalizationData.l1RollingHashMessageNumber,
    l2MerkleRoots: finalizationData.l2MerkleRoots,
    l2MerkleTreesDepth: finalizationData.l2MerkleTreesDepth,
    l2MessagingBlocksOffsets: finalizationData.l2MessagingBlocksOffsets,
  };
}

export function generateCallDataSubmission(startDataIndex: number, finalDataIndex: number): CalldataSubmissionData[] {
  return COMPRESSED_SUBMISSION_DATA.slice(startDataIndex, finalDataIndex).map((data) => {
    const returnData = {
      parentStateRootHash: data.parentStateRootHash,
      dataParentHash: data.parentDataHash,
      finalStateRootHash: data.finalStateRootHash,
      firstBlockInData: BigInt(data.conflationOrder.startingBlockNumber),
      finalBlockInData: BigInt(data.conflationOrder.upperBoundaries.slice(-1)[0]),
      snarkHash: data.snarkHash,
      compressedData: ethers.hexlify(ethers.decodeBase64(data.compressedData)),
    };
    return returnData;
  });
}

export function generateSubmissionData(startDataIndex: number, finalDataIndex: number): SubmissionAndCompressedData[] {
  return COMPRESSED_SUBMISSION_DATA.slice(startDataIndex, finalDataIndex).map((data) => {
    return {
      submissionData: {
        parentStateRootHash: data.parentStateRootHash,
        dataParentHash: data.parentDataHash,
        finalStateRootHash: data.finalStateRootHash,
        firstBlockInData: BigInt(data.conflationOrder.startingBlockNumber),
        finalBlockInData: BigInt(data.conflationOrder.upperBoundaries.slice(-1)[0]),
        snarkHash: data.snarkHash,
      },
      compressedData: ethers.hexlify(ethers.decodeBase64(data.compressedData)),
    };
  });
}

//TODO Refactor
export function generateSubmissionDataMultipleProofs(
  startDataIndex: number,
  finalDataIndex: number,
): SubmissionAndCompressedData[] {
  return COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF.slice(startDataIndex, finalDataIndex).map((data) => {
    return {
      submissionData: {
        parentStateRootHash: data.parentStateRootHash,
        dataParentHash: data.parentDataHash,
        finalStateRootHash: data.finalStateRootHash,
        firstBlockInData: BigInt(data.conflationOrder.startingBlockNumber),
        finalBlockInData: BigInt(data.conflationOrder.upperBoundaries.slice(-1)[0]),
        snarkHash: data.snarkHash,
      },
      compressedData: ethers.hexlify(ethers.decodeBase64(data.compressedData)),
    };
  });
}

export function generateCallDataSubmissionMultipleProofs(
  startDataIndex: number,
  finalDataIndex: number,
): CalldataSubmissionData[] {
  return COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF.slice(startDataIndex, finalDataIndex).map((data) => {
    const returnData = {
      parentStateRootHash: data.parentStateRootHash,
      dataParentHash: data.parentDataHash,
      finalStateRootHash: data.finalStateRootHash,
      firstBlockInData: BigInt(data.conflationOrder.startingBlockNumber),
      finalBlockInData: BigInt(data.conflationOrder.upperBoundaries.slice(-1)[0]),
      snarkHash: data.snarkHash,
      compressedData: ethers.hexlify(ethers.decodeBase64(data.compressedData)),
    };
    return returnData;
  });
}

export function generateSubmissionDataFromJSON(
  startingBlockNumber: number,
  endingBlockNumber: number,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  parsedJSONData: any,
): SubmissionData {
  const returnData = {
    parentStateRootHash: parsedJSONData.parentStateRootHash,
    dataParentHash: parsedJSONData.parentDataHash,
    finalStateRootHash: parsedJSONData.finalStateRootHash,
    firstBlockInData: BigInt(startingBlockNumber),
    finalBlockInData: BigInt(endingBlockNumber),
    snarkHash: parsedJSONData.snarkHash,
    compressedData: ethers.hexlify(ethers.decodeBase64(parsedJSONData.compressedData)),
  };

  return returnData;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function generateFinalizationDataFromJSON(parsedJSONData: any): FinalizationData {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { aggregatedProverVersion, aggregatedVerifierIndex, aggregatedProofPublicInput, ...data } = parsedJSONData;
  return {
    ...data,
    finalBlockNumber: BigInt(data.finalBlockNumber),
    l1RollingHashMessageNumber: BigInt(data.l1RollingHashMessageNumber),
    l2MerkleTreesDepth: BigInt(data.l2MerkleTreesDepth),
    l2MessagingBlocksOffsets: data.l2MessagingBlocksOffsets,
  };
}
