import { ethers } from "ethers";
import { time as networkTime } from "@nomicfoundation/hardhat-network-helpers";
import {
  HASH_ZERO,
  COMPRESSED_SUBMISSION_DATA,
  COMPRESSED_SUBMISSION_DATA_MULTIPLE_PROOF,
  BLOB_SUBMISSION_DATA,
  BLOB_SUBMISSION_DATA_MULTIPLE_PROOF,
} from "../constants";
import {
  FinalizationData,
  CalldataSubmissionData,
  SubmissionAndCompressedData,
  BlobSubmissionData,
  ShnarfData,
  ParentSubmissionData,
  ParentAndExpectedShnarf,
  SubmissionData,
} from "../types";
import { generateRandomBytes, range } from "./general";
import { generateKeccak256 } from "./hashing";

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
