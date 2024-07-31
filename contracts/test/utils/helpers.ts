import { time as networkTime } from "@nomicfoundation/hardhat-network-helpers";
import fs from "fs";
import path from "path";
import { expect } from "chai";
import { DEFAULT_MESSAGE_NONCE, HASH_ZERO, MESSAGE_FEE, MESSAGE_VALUE_1ETH } from "./constants";
import {
  DebugData,
  FormattedBlockData,
  RawBlockData,
  FinalizationData,
  SubmissionData,
  SubmissionAndCompressedData,
  CalldataSubmissionData,
  ParentSubmissionData,
  ShnarfData,
  ParentAndExpectedShnarf,
  BlobSubmissionData,
} from "./types";

import firstCompressedDataContent from "../testData/compressedData/blocks-1-46.json";
import fourthCompressedDataContent from "../testData/compressedData/blocks-115-155.json";
import secondCompressedDataContent from "../testData/compressedData/blocks-47-81.json";
import thirdCompressedDataContent from "../testData/compressedData/blocks-82-114.json";
import fourthBlobDataContent from "../testData/compressedDataEip4844/blocks-115-155.json";
import thirdBlobDataContent from "../testData/compressedDataEip4844/blocks-82-114.json";
import secondBlobDataContent from "../testData/compressedDataEip4844/blocks-47-81.json";
import firstBlobDataContent from "../testData/compressedDataEip4844/blocks-1-46.json";

const COMPRESSED_SUBMISSION_DATA = [
  firstCompressedDataContent,
  secondCompressedDataContent,
  thirdCompressedDataContent,
  fourthCompressedDataContent,
];

const BLOB_SUBMISSION_DATA = [firstBlobDataContent, secondBlobDataContent, thirdBlobDataContent, fourthBlobDataContent];

import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { AbiCoder, ethers } from "ethers";
import firstCompressedDataContent_multiple from "../testData/compressedData/multipleProofs/blocks-1-46.json";
import secondCompressedDataContent_multiple from "../testData/compressedData/multipleProofs/blocks-47-81.json";
import thirdCompressedDataContent_multiple from "../testData/compressedData/multipleProofs/blocks-82-119.json";
import fourthCompressedDataContent_multiple from "../testData/compressedData/multipleProofs/blocks-120-153.json";

const COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF = [
  firstCompressedDataContent_multiple,
  secondCompressedDataContent_multiple,
  thirdCompressedDataContent_multiple,
  fourthCompressedDataContent_multiple,
];

import firstBlobContent_multiple from "../testData/compressedDataEip4844/multipleProofs/blocks-1-46.json";
import secondBlobContent_multiple from "../testData/compressedDataEip4844/multipleProofs/blocks-47-81.json";
import thirdBlobContent_multiple from "../testData/compressedDataEip4844/multipleProofs/blocks-82-119.json";
import fourthBlobContent_multiple from "../testData/compressedDataEip4844/multipleProofs/blocks-120-153.json";

const BLOB_SUBMISSION_DATA_MULTIPLE_PROOF = [
  firstBlobContent_multiple,
  secondBlobContent_multiple,
  thirdBlobContent_multiple,
  fourthBlobContent_multiple,
];
export type ShnarfDataGenerator = (blobParentShnarfIndex: number, isMultiple?: boolean) => ShnarfData;

export const generateKeccak256Hash = (str: string) => generateKeccak256(["string"], [str], true);

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
    finalBlockInData: 99n,
    lastFinalizedShnarf: generateParentSubmissionDataForIndex(1).shnarf,
    shnarfData: generateParentShnarfData(1),
    parentStateRootHash: generateRandomBytes(32),
    lastFinalizedTimestamp: BigInt((await networkTime.latest()) - 2),
    finalTimestamp: BigInt(await networkTime.latest()),
    l1RollingHash: generateRandomBytes(32),
    l1RollingHashMessageNumber: 10n,
    l2MerkleRoots: [generateRandomBytes(32)],
    l2MerkleTreesDepth: 5n,
    l2MessagingBlocksOffsets: generateL2MessagingBlocksOffsets(1, 1),
    lastFinalizedL1RollingHash: HASH_ZERO,
    lastFinalizedL1RollingHashMessageNumber: 0n,
    ...overrides,
  };
}

export function generateCallDataSubmission(startDataIndex: number, finalDataIndex: number): CalldataSubmissionData[] {
  return COMPRESSED_SUBMISSION_DATA.slice(startDataIndex, finalDataIndex).map((data) => {
    const returnData = {
      finalStateRootHash: data.finalStateRootHash,
      firstBlockInData: BigInt(data.conflationOrder.startingBlockNumber),
      finalBlockInData: BigInt(data.conflationOrder.upperBoundaries.slice(-1)[0]),
      snarkHash: data.snarkHash,
      compressedData: ethers.hexlify(ethers.decodeBase64(data.compressedData)),
    };
    return returnData;
  });
}

export function generateBlobDataSubmission(
  startDataIndex: number,
  finalDataIndex: number,
  isMultiple: boolean = false,
): { blobDataSubmission: BlobSubmissionData[]; compressedBlobs: string[]; parentShnarf: string; finalShnarf: string } {
  const dataSet = isMultiple ? BLOB_SUBMISSION_DATA_MULTIPLE_PROOF : BLOB_SUBMISSION_DATA;
  const compressedBlobs: string[] = [];
  const parentShnarf = dataSet[startDataIndex].prevShnarf;
  const finalShnarf = dataSet[finalDataIndex - 1].expectedShnarf;
  const blobDataSubmission = dataSet.slice(startDataIndex, finalDataIndex).map((data) => {
    compressedBlobs.push(ethers.hexlify(ethers.decodeBase64(data.compressedData)));
    const returnData: BlobSubmissionData = {
      submissionData: {
        finalStateRootHash: data.finalStateRootHash,
        firstBlockInData: BigInt(data.conflationOrder.startingBlockNumber),
        finalBlockInData: BigInt(data.conflationOrder.upperBoundaries.slice(-1)[0]),
        snarkHash: data.snarkHash,
      },
      dataEvaluationClaim: data.expectedY,
      kzgCommitment: data.commitment,
      kzgProof: data.kzgProofContract,
    };
    return returnData;
  });
  return {
    compressedBlobs,
    blobDataSubmission,
    parentShnarf,
    finalShnarf,
  };
}

export function generateParentShnarfData(index: number, multiple?: boolean): ShnarfData {
  if (index === 0) {
    return {
      parentShnarf: HASH_ZERO,
      snarkHash: HASH_ZERO,
      finalStateRootHash: multiple
        ? COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF[0].parentStateRootHash
        : COMPRESSED_SUBMISSION_DATA[0].parentStateRootHash,
      dataEvaluationPoint: HASH_ZERO,
      dataEvaluationClaim: HASH_ZERO,
    };
  }
  const parentSubmissionData = multiple
    ? COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF[index - 1]
    : COMPRESSED_SUBMISSION_DATA[index - 1];

  return {
    parentShnarf: parentSubmissionData.prevShnarf,
    snarkHash: parentSubmissionData.snarkHash,
    finalStateRootHash: parentSubmissionData.finalStateRootHash,
    dataEvaluationPoint: parentSubmissionData.expectedX,
    dataEvaluationClaim: parentSubmissionData.expectedY,
  };
}

export function generateBlobParentShnarfData(index: number, multiple?: boolean): ShnarfData {
  if (index === 0) {
    return {
      parentShnarf: HASH_ZERO,
      snarkHash: HASH_ZERO,
      finalStateRootHash: multiple
        ? BLOB_SUBMISSION_DATA_MULTIPLE_PROOF[0].parentStateRootHash
        : BLOB_SUBMISSION_DATA[0].parentStateRootHash,
      dataEvaluationPoint: HASH_ZERO,
      dataEvaluationClaim: HASH_ZERO,
    };
  }
  const parentSubmissionData = multiple
    ? BLOB_SUBMISSION_DATA_MULTIPLE_PROOF[index - 1]
    : BLOB_SUBMISSION_DATA[index - 1];

  return {
    parentShnarf: parentSubmissionData.prevShnarf,
    snarkHash: parentSubmissionData.snarkHash,
    finalStateRootHash: parentSubmissionData.finalStateRootHash,
    dataEvaluationPoint: parentSubmissionData.expectedX,
    dataEvaluationClaim: parentSubmissionData.expectedY,
  };
}

export function generateExpectedParentSubmissionHash(
  firstBlockInData: bigint,
  finalBlockInData: bigint,
  finalStateRootHash: string,
  shnarf: string,
  dataParentHash: string,
): string {
  return generateKeccak256(
    ["uint256", "uint256", "bytes32", "bytes32", "bytes32"],
    [firstBlockInData, finalBlockInData, finalStateRootHash, shnarf, dataParentHash],
  );
}

export function generateParentSubmissionDataForIndex(index: number): ParentSubmissionData {
  if (index === 0) {
    return {
      finalStateRootHash: COMPRESSED_SUBMISSION_DATA[0].parentStateRootHash,
      firstBlockInData: 0n,
      finalBlockInData: 0n,
      shnarf: generateKeccak256(
        ["bytes32", "bytes32", "bytes32", "bytes32", "bytes32"],
        [HASH_ZERO, HASH_ZERO, COMPRESSED_SUBMISSION_DATA[0].parentStateRootHash, HASH_ZERO, HASH_ZERO],
      ),
    };
  }

  return {
    finalStateRootHash: COMPRESSED_SUBMISSION_DATA[index - 1].finalStateRootHash,
    firstBlockInData: BigInt(COMPRESSED_SUBMISSION_DATA[index - 1].conflationOrder.startingBlockNumber),
    finalBlockInData: BigInt(COMPRESSED_SUBMISSION_DATA[index - 1].conflationOrder.upperBoundaries.slice(-1)[0]),
    shnarf: COMPRESSED_SUBMISSION_DATA[index - 1].expectedShnarf,
  };
}

export function generateParentAndExpectedShnarfForIndex(index: number): ParentAndExpectedShnarf {
  return {
    parentShnarf: COMPRESSED_SUBMISSION_DATA[index].prevShnarf,
    expectedShnarf: COMPRESSED_SUBMISSION_DATA[index].expectedShnarf,
  };
}

export function generateParentAndExpectedShnarfForMulitpleIndex(index: number): ParentAndExpectedShnarf {
  return {
    parentShnarf: COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF[index].prevShnarf,
    expectedShnarf: COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF[index].expectedShnarf,
  };
}

export function generateParentSubmissionDataForIndexForMultiple(index: number): ParentSubmissionData {
  if (index === 0) {
    return {
      finalStateRootHash: COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF[0].parentStateRootHash,
      firstBlockInData: 0n,
      finalBlockInData: 0n,
      shnarf: generateKeccak256(
        ["bytes32", "bytes32", "bytes32", "bytes32", "bytes32"],
        [HASH_ZERO, HASH_ZERO, COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF[0].parentStateRootHash, HASH_ZERO, HASH_ZERO],
      ),
    };
  }
  return {
    finalStateRootHash: COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF[index - 1].finalStateRootHash,
    firstBlockInData: BigInt(COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF[index - 1].conflationOrder.startingBlockNumber),
    finalBlockInData: BigInt(
      COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF[index - 1].conflationOrder.upperBoundaries.slice(-1)[0],
    ),
    shnarf: COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF[index - 1].expectedShnarf,
  };
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
      parentShnarf: data.prevShnarf,
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

/* eslint-disable @typescript-eslint/no-explicit-any */
export async function expectRevertWithCustomError(
  contract: any,
  asyncCall: Promise<any>,
  errorName: string,
  errorArgs: any[] = [],
) {
  await expect(asyncCall)
    .to.be.revertedWithCustomError(contract, errorName)
    .withArgs(...errorArgs);
}

/* eslint-disable @typescript-eslint/no-explicit-any */
export async function expectRevertWithReason(asyncCall: Promise<any>, reason: string) {
  await expect(asyncCall).to.be.revertedWith(reason);
}
export function buildAccessErrorMessage(account: SignerWithAddress, role: string): string {
  return `AccessControl: account ${account.address.toLowerCase()} is missing role ${role}`;
}

/* eslint-disable @typescript-eslint/no-explicit-any */
export async function expectEvent(contract: any, asyncCall: Promise<any>, eventName: string, eventArgs: any[] = []) {
  await expect(asyncCall)
    .to.emit(contract, eventName)
    .withArgs(...eventArgs);
}
