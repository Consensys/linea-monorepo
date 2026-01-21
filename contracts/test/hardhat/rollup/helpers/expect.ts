import { expect } from "chai";
import { BaseContract, Contract, Transaction } from "ethers";
import { ethers } from "hardhat";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";

import { TestLineaRollup } from "contracts/typechain-types";
import { getWalletForIndex } from "./";
import { HASH_ZERO, TEST_PUBLIC_VERIFIER_INDEX } from "../../common/constants";
import {
  expectEvent,
  expectEventDirectFromReceiptData,
  expectRevertWithCustomError,
  generateFinalizationData,
  generateKeccak256,
} from "../../common/helpers";
import { ShnarfDataGenerator } from "../../common/types";

export async function expectSuccessfulFinalize(
  lineaRollup: TestLineaRollup,
  operator: SignerWithAddress,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  proofData: any,
  blobParentShnarfIndex: number,
  finalStateRootHash: string,
  shnarfDataGenerator: ShnarfDataGenerator,
  isMultiple: boolean = false,
  lastFinalizedRollingHash: string = HASH_ZERO,
  lastFinalizedMessageNumber: bigint = 0n,
  lastFinalizedForcedTransactionRollingHash: string = HASH_ZERO,
  lastFinalizedForcedTransactionNumber: bigint = 0n,
  finalForcedTransactionNumber: bigint = 0n,
  lastFinalizedBlockHash: string = HASH_ZERO,
  finalBlockHash: string = HASH_ZERO,
  filteredAddresses: string[] = [],
) {
  const finalizationData = await generateFinalizationData({
    l1RollingHash: proofData.l1RollingHash,
    l1RollingHashMessageNumber: BigInt(proofData.l1RollingHashMessageNumber),
    lastFinalizedTimestamp: BigInt(proofData.parentAggregationLastBlockTimestamp),
    endBlockNumber: BigInt(proofData.finalBlockNumber),
    parentStateRootHash: proofData.parentStateRootHash,
    finalTimestamp: BigInt(proofData.finalTimestamp),
    l2MerkleRoots: proofData.l2MerkleRoots,
    l2MerkleTreesDepth: BigInt(proofData.l2MerkleTreesDepth),
    l2MessagingBlocksOffsets: proofData.l2MessagingBlocksOffsets,
    filteredAddresses: filteredAddresses,
    aggregatedProof: proofData.aggregatedProof,
    shnarfData: shnarfDataGenerator(blobParentShnarfIndex, isMultiple),
  });

  finalizationData.lastFinalizedL1RollingHash = lastFinalizedRollingHash;
  finalizationData.lastFinalizedL1RollingHashMessageNumber = lastFinalizedMessageNumber;

  finalizationData.lastFinalizedForcedTransactionRollingHash = lastFinalizedForcedTransactionRollingHash;
  finalizationData.lastFinalizedForcedTransactionNumber = lastFinalizedForcedTransactionNumber;
  finalizationData.finalForcedTransactionNumber = finalForcedTransactionNumber;
  finalizationData.lastFinalizedBlockHash = lastFinalizedBlockHash;
  finalizationData.finalBlockHash = finalBlockHash;

  await lineaRollup.setRollingHash(proofData.l1RollingHashMessageNumber, proofData.l1RollingHash);

  const finalizeCompressedCall = lineaRollup
    .connect(operator)
    .finalizeBlocks(proofData.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);

  const eventArgs = [
    BigInt(proofData.lastFinalizedBlockNumber) + 1n,
    finalizationData.endBlockNumber,
    proofData.finalShnarf,
    finalizationData.parentStateRootHash,
    finalStateRootHash,
  ];

  await expectEvent(lineaRollup, finalizeCompressedCall, "DataFinalizedV3", eventArgs);

  const [expectedFinalStateRootHash, lastFinalizedBlockNumber, lastFinalizedState] = await Promise.all([
    lineaRollup.stateRootHashes(finalizationData.endBlockNumber),
    lineaRollup.currentL2BlockNumber(),
    lineaRollup.currentFinalizedState(),
  ]);

  expect(expectedFinalStateRootHash).to.equal(finalizationData.shnarfData.finalStateRootHash);
  expect(lastFinalizedBlockNumber).to.equal(finalizationData.endBlockNumber);
  expect(lastFinalizedState).to.equal(
    generateKeccak256(
      ["uint256", "bytes32", "uint256", "bytes32", "uint256", "bytes32"],
      [
        finalizationData.l1RollingHashMessageNumber,
        finalizationData.l1RollingHash,
        finalForcedTransactionNumber,
        proofData.finalFtxRollingHash,
        finalizationData.finalTimestamp,
        finalizationData.finalBlockHash,
      ],
    ),
  );
}

export async function expectFailedCustomErrorFinalize(
  lineaRollup: TestLineaRollup,
  operator: SignerWithAddress,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  proofData: any,
  blobParentShnarfIndex: number,
  finalStateRootHash: string,
  shnarfDataGenerator: ShnarfDataGenerator,
  isMultiple: boolean = false,
  lastFinalizedRollingHash: string = HASH_ZERO,
  lastFinalizedMessageNumber: bigint = 0n,
  lastFinalizedForcedTransactionRollingHash: string = HASH_ZERO,
  lastFinalizedForcedTransactionNumber: bigint = 0n,
  filteredAddresses: string[] = [],
  customError: string,
  customErrorArgs: unknown[] = [],
) {
  const finalizationData = await generateFinalizationData({
    l1RollingHash: proofData.l1RollingHash,
    l1RollingHashMessageNumber: BigInt(proofData.l1RollingHashMessageNumber),
    lastFinalizedTimestamp: BigInt(proofData.parentAggregationLastBlockTimestamp),
    endBlockNumber: BigInt(proofData.finalBlockNumber),
    parentStateRootHash: proofData.parentStateRootHash,
    finalTimestamp: BigInt(proofData.finalTimestamp),
    l2MerkleRoots: proofData.l2MerkleRoots,
    l2MerkleTreesDepth: BigInt(proofData.l2MerkleTreesDepth),
    l2MessagingBlocksOffsets: proofData.l2MessagingBlocksOffsets,
    filteredAddresses: filteredAddresses,
    aggregatedProof: proofData.aggregatedProof,
    shnarfData: shnarfDataGenerator(blobParentShnarfIndex, isMultiple),
  });

  finalizationData.lastFinalizedL1RollingHash = lastFinalizedRollingHash;
  finalizationData.lastFinalizedL1RollingHashMessageNumber = lastFinalizedMessageNumber;
  finalizationData.lastFinalizedForcedTransactionRollingHash = lastFinalizedForcedTransactionRollingHash;
  finalizationData.lastFinalizedForcedTransactionNumber = lastFinalizedForcedTransactionNumber;

  await lineaRollup.setRollingHash(proofData.l1RollingHashMessageNumber, proofData.l1RollingHash);

  const finalizeCompressedCall = lineaRollup
    .connect(operator)
    .finalizeBlocks(proofData.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);

  await expectRevertWithCustomError(lineaRollup, finalizeCompressedCall, customError, customErrorArgs);
}

export async function expectSuccessfulFinalizeViaCallForwarder(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  proofData: any,
  blobParentShnarfIndex: number,
  finalStateRootHash: string,
  shnarfDataGenerator: ShnarfDataGenerator,
  isMultiple: boolean = false,
  lastFinalizedRollingHash: string = HASH_ZERO,
  lastFinalizedMessageNumber: bigint = 0n,
  callforwarderAddress: string,
  upgradedContract: Contract,
) {
  const finalizationData = await generateFinalizationData({
    l1RollingHash: proofData.l1RollingHash,
    l1RollingHashMessageNumber: BigInt(proofData.l1RollingHashMessageNumber),
    lastFinalizedTimestamp: BigInt(proofData.parentAggregationLastBlockTimestamp),
    endBlockNumber: BigInt(proofData.finalBlockNumber),
    parentStateRootHash: proofData.parentStateRootHash,
    finalTimestamp: BigInt(proofData.finalTimestamp),
    l2MerkleRoots: proofData.l2MerkleRoots,
    l2MerkleTreesDepth: BigInt(proofData.l2MerkleTreesDepth),
    l2MessagingBlocksOffsets: proofData.l2MessagingBlocksOffsets,
    aggregatedProof: proofData.aggregatedProof,
    shnarfData: shnarfDataGenerator(blobParentShnarfIndex, isMultiple),
  });
  finalizationData.lastFinalizedL1RollingHash = lastFinalizedRollingHash;
  finalizationData.lastFinalizedL1RollingHashMessageNumber = lastFinalizedMessageNumber;

  await upgradedContract.setRollingHash(proofData.l1RollingHashMessageNumber, proofData.l1RollingHash);

  const shnarfData = shnarfDataGenerator(blobParentShnarfIndex, isMultiple);

  const finalShnarf = generateKeccak256(
    ["bytes32", "bytes32", "bytes32", "bytes32", "bytes32"],
    [
      shnarfData.parentShnarf,
      shnarfData.snarkHash,
      shnarfData.finalStateRootHash,
      shnarfData.dataEvaluationPoint,
      shnarfData.dataEvaluationClaim,
    ],
  );
  const blobShnarfExists = await upgradedContract.blobShnarfExists(finalShnarf);
  expect(blobShnarfExists).to.equal(1n);

  await upgradedContract.setRollingHash(proofData.l1RollingHashMessageNumber, proofData.l1RollingHash);

  const txData = [
    proofData.aggregatedProof,
    0,
    [
      proofData.parentStateRootHash,
      BigInt(proofData.finalBlockNumber),
      [
        shnarfData.parentShnarf,
        shnarfData.snarkHash,
        shnarfData.finalStateRootHash,
        shnarfData.dataEvaluationPoint,
        shnarfData.dataEvaluationClaim,
      ],
      proofData.parentAggregationLastBlockTimestamp,
      proofData.finalTimestamp,
      lastFinalizedRollingHash,
      proofData.l1RollingHash,
      lastFinalizedMessageNumber,
      proofData.l1RollingHashMessageNumber,
      proofData.l2MerkleTreesDepth,
      finalizationData.lastFinalizedForcedTransactionNumber,
      finalizationData.finalForcedTransactionNumber,
      finalizationData.lastFinalizedForcedTransactionRollingHash,
      proofData.l2MerkleRoots,
      [],
      proofData.l2MessagingBlocksOffsets,
    ],
  ];

  const encodedCall = ethers.concat([
    "0x755bc62f",
    ethers.AbiCoder.defaultAbiCoder().encode(
      [
        "bytes",
        "uint256",
        "tuple(bytes32,uint256,tuple(bytes32,bytes32,bytes32,bytes32,bytes32),uint256,uint256,bytes32,bytes32,uint256,uint256,uint256,uint256,uint256,bytes32,bytes32[],address[],bytes)",
      ],
      txData,
    ),
  ]);

  const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
  const operatorHDSigner = getWalletForIndex(2);
  const nonce = await operatorHDSigner.getNonce();

  const transaction = Transaction.from({
    data: encodedCall,
    maxPriorityFeePerGas: maxPriorityFeePerGas!,
    maxFeePerGas: maxFeePerGas!,
    to: callforwarderAddress,
    chainId: (await ethers.provider.getNetwork()).chainId,
    type: 2,
    nonce: nonce,
    value: 0,
    gasLimit: 10_000_000,
  });

  const signedTx = await operatorHDSigner.signTransaction(transaction);

  const txResponse = await ethers.provider.broadcastTransaction(signedTx);
  const receipt = await ethers.provider.getTransactionReceipt(txResponse.hash);
  expect(receipt).is.not.null;

  const eventArgs = [
    BigInt(proofData.lastFinalizedBlockNumber) + 1n,
    finalizationData.endBlockNumber,
    proofData.finalShnarf,
    finalizationData.parentStateRootHash,
    finalStateRootHash,
  ];

  const dataFinalizedLogIndex = 8;

  expectEventDirectFromReceiptData(
    upgradedContract as BaseContract,
    receipt!,
    "DataFinalizedV3",
    eventArgs,
    dataFinalizedLogIndex,
  );

  const [expectedFinalStateRootHash, lastFinalizedBlockNumber, lastFinalizedState] = await Promise.all([
    upgradedContract.stateRootHashes(finalizationData.endBlockNumber),
    upgradedContract.currentL2BlockNumber(),
    upgradedContract.currentFinalizedState(),
  ]);

  expect(expectedFinalStateRootHash).to.equal(finalizationData.shnarfData.finalStateRootHash);
  expect(lastFinalizedBlockNumber).to.equal(finalizationData.endBlockNumber);
  expect(lastFinalizedState).to.equal(
    generateKeccak256(
      ["uint256", "bytes32", "uint256", "bytes32", "uint256"],
      [
        finalizationData.l1RollingHashMessageNumber,
        finalizationData.l1RollingHash,
        0,
        HASH_ZERO,
        finalizationData.finalTimestamp,
      ],
    ),
  );
}
