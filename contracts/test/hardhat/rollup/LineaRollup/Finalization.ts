import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture, time as networkTime } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";

import aggregatedProof1To81 from "../../_testData/compressedData/multipleProofs/aggregatedProof-1-81.json";
import aggregatedProof82To153 from "../../_testData/compressedData/multipleProofs/aggregatedProof-82-153.json";
import calldataAggregatedProof1To155 from "../../_testData/compressedData/aggregatedProof-1-155.json";
import secondCompressedDataContent from "../../_testData/compressedData/blocks-47-81.json";
import fourthCompressedDataContent from "../../_testData/compressedData/blocks-115-155.json";
import fourthMultipleCompressedDataContent from "../../_testData/compressedData/multipleProofs/blocks-120-153.json";

import { LineaRollup__factory, TestLineaRollup } from "contracts/typechain-types";
import { expectSuccessfulFinalize, getAccountsFixture, deployLineaRollupFixture } from "./../helpers";
import {
  GENERAL_PAUSE_TYPE,
  HASH_ZERO,
  OPERATOR_ROLE,
  TEST_PUBLIC_VERIFIER_INDEX,
  EMPTY_CALLDATA,
  FINALIZATION_PAUSE_TYPE,
  DEFAULT_LAST_FINALIZED_TIMESTAMP,
  MAX_GAS_LIMIT,
} from "../../common/constants";
import {
  calculateRollingHash,
  generateFinalizationData,
  generateRandomBytes,
  generateCallDataSubmission,
  generateCallDataSubmissionMultipleProofs,
  generateKeccak256,
  buildAccessErrorMessage,
  expectRevertWithCustomError,
  expectRevertWithReason,
  generateParentAndExpectedShnarfForIndex,
  generateParentAndExpectedShnarfForMulitpleIndex,
  generateParentShnarfData,
} from "../../common/helpers";
import { expect } from "chai";

describe("Linea Rollup contract: Finalization", () => {
  let lineaRollup: TestLineaRollup;

  let securityCouncil: SignerWithAddress;
  let operator: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;

  before(async () => {
    ({ securityCouncil, operator, nonAuthorizedAccount } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ lineaRollup } = await loadFixture(deployLineaRollupFixture));
  });

  describe("Blocks finalization with proof", () => {
    const messageHash = generateRandomBytes(32);

    beforeEach(async () => {
      await lineaRollup.addRollingHash(10, messageHash);
      await lineaRollup.setLastFinalizedBlock(0);
    });

    describe("With and without submission data", () => {
      it("Should revert if l1 message number == 0 and l1 rolling hash is not empty", async () => {
        const finalizationData = await generateFinalizationData({
          l1RollingHashMessageNumber: 0n,
          l1RollingHash: generateRandomBytes(32),
        });

        const lastFinalizedBlockNumber = await lineaRollup.currentL2BlockNumber();
        const parentStateRootHash = await lineaRollup.stateRootHashes(lastFinalizedBlockNumber);
        finalizationData.parentStateRootHash = parentStateRootHash;

        const proof = calldataAggregatedProof1To155.aggregatedProof;

        const finalizeCall = lineaRollup
          .connect(operator)
          .finalizeBlocks(proof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCall, "MissingMessageNumberForRollingHash", [
          finalizationData.l1RollingHash,
        ]);
      });

      it("Should revert if l1 message number != 0 and l1 rolling hash is empty", async () => {
        const finalizationData = await generateFinalizationData({
          l1RollingHashMessageNumber: 1n,
          l1RollingHash: HASH_ZERO,
        });

        const lastFinalizedBlockNumber = await lineaRollup.currentL2BlockNumber();
        const parentStateRootHash = await lineaRollup.stateRootHashes(lastFinalizedBlockNumber);
        finalizationData.parentStateRootHash = parentStateRootHash;

        const proof = calldataAggregatedProof1To155.aggregatedProof;

        const finalizeCall = lineaRollup
          .connect(operator)
          .finalizeBlocks(proof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCall, "MissingRollingHashForMessageNumber", [
          finalizationData.l1RollingHashMessageNumber,
        ]);
      });

      it("Should revert if l1RollingHash does not exist on L1", async () => {
        const finalizationData = await generateFinalizationData({
          l1RollingHashMessageNumber: 1n,
          l1RollingHash: generateRandomBytes(32),
        });

        const lastFinalizedBlockNumber = await lineaRollup.currentL2BlockNumber();
        const parentStateRootHash = await lineaRollup.stateRootHashes(lastFinalizedBlockNumber);
        finalizationData.parentStateRootHash = parentStateRootHash;

        const proof = calldataAggregatedProof1To155.aggregatedProof;

        const finalizeCall = lineaRollup
          .connect(operator)
          .finalizeBlocks(proof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCall, "L1RollingHashDoesNotExistOnL1", [
          finalizationData.l1RollingHashMessageNumber,
          finalizationData.l1RollingHash,
        ]);
      });

      it("Should revert if timestamps are not in sequence", async () => {
        const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);
        let index = 0;
        for (const data of submissionDataBeforeFinalization) {
          const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
          await lineaRollup
            .connect(operator)
            .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
              gasLimit: MAX_GAS_LIMIT,
            });
          index++;
        }

        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
          lastFinalizedTimestamp: DEFAULT_LAST_FINALIZED_TIMESTAMP,
          endBlockNumber: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
          parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
          finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
          l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
          l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
          l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
          aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
          shnarfData: generateParentShnarfData(index),
        });

        await lineaRollup.setRollingHash(
          calldataAggregatedProof1To155.l1RollingHashMessageNumber,
          calldataAggregatedProof1To155.l1RollingHash,
        );

        finalizationData.lastFinalizedTimestamp = finalizationData.finalTimestamp + 1n;

        const expectedHashValue = generateKeccak256(
          ["uint256", "bytes32", "uint256"],
          [
            finalizationData.lastFinalizedL1RollingHashMessageNumber,
            finalizationData.lastFinalizedL1RollingHash,
            finalizationData.lastFinalizedTimestamp,
          ],
        );
        const actualHashValue = generateKeccak256(
          ["uint256", "bytes32", "uint256"],
          [
            finalizationData.lastFinalizedL1RollingHashMessageNumber,
            finalizationData.lastFinalizedL1RollingHash,
            DEFAULT_LAST_FINALIZED_TIMESTAMP,
          ],
        );

        const finalizeCompressedCall = lineaRollup
          .connect(operator)
          .finalizeBlocks(calldataAggregatedProof1To155.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCompressedCall, "FinalizationStateIncorrect", [
          expectedHashValue,
          actualHashValue,
        ]);
      });

      it("Should revert if the final shnarf does not exist", async () => {
        const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);
        let index = 0;
        for (const data of submissionDataBeforeFinalization) {
          const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
          await lineaRollup
            .connect(operator)
            .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
              gasLimit: MAX_GAS_LIMIT,
            });
          index++;
        }

        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
          lastFinalizedTimestamp: DEFAULT_LAST_FINALIZED_TIMESTAMP,
          endBlockNumber: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
          parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
          finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
          l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
          l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
          l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
          aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
          shnarfData: generateParentShnarfData(index),
        });

        await lineaRollup.setRollingHash(
          calldataAggregatedProof1To155.l1RollingHashMessageNumber,
          calldataAggregatedProof1To155.l1RollingHash,
        );

        finalizationData.shnarfData.snarkHash = generateRandomBytes(32);

        const { dataEvaluationClaim, dataEvaluationPoint, finalStateRootHash, parentShnarf, snarkHash } =
          finalizationData.shnarfData;
        const expectedMissingBlobShnarf = generateKeccak256(
          ["bytes32", "bytes32", "bytes32", "bytes32", "bytes32"],
          [parentShnarf, snarkHash, finalStateRootHash, dataEvaluationPoint, dataEvaluationClaim],
        );

        const finalizeCompressedCall = lineaRollup
          .connect(operator)
          .finalizeBlocks(calldataAggregatedProof1To155.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCompressedCall, "FinalShnarfNotSubmitted", [
          expectedMissingBlobShnarf,
        ]);
      });

      it("Should revert if finalizationData.finalTimestamp is greater than the block.timestamp", async () => {
        const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);
        let index = 0;
        for (const data of submissionDataBeforeFinalization) {
          const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
          await lineaRollup
            .connect(operator)
            .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
              gasLimit: MAX_GAS_LIMIT,
            });
          index++;
        }

        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
          lastFinalizedTimestamp: DEFAULT_LAST_FINALIZED_TIMESTAMP,
          endBlockNumber: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
          parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
          finalTimestamp: BigInt(new Date(new Date().setHours(new Date().getHours() + 2)).getTime()), // Set to 2 hours in the future
          l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
          l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
          l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
          aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
          shnarfData: generateParentShnarfData(index),
        });

        await lineaRollup.setRollingHash(
          calldataAggregatedProof1To155.l1RollingHashMessageNumber,
          calldataAggregatedProof1To155.l1RollingHash,
        );

        const finalizeCompressedCall = lineaRollup
          .connect(operator)
          .finalizeBlocks(calldataAggregatedProof1To155.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCompressedCall, "FinalizationInTheFuture", [
          finalizationData.finalTimestamp,
          (await networkTime.latest()) + 1,
        ]);
      });
    });

    describe("Without submission data", () => {
      it("Should revert if the final block state equals the zero hash", async () => {
        const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);
        let index = 0;
        for (const data of submissionDataBeforeFinalization) {
          const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
          await lineaRollup
            .connect(operator)
            .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
              gasLimit: MAX_GAS_LIMIT,
            });
          index++;
        }

        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
          lastFinalizedTimestamp: DEFAULT_LAST_FINALIZED_TIMESTAMP,
          endBlockNumber: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
          parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
          finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
          l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
          l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
          l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
          aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
          shnarfData: generateParentShnarfData(index),
        });

        await lineaRollup.setRollingHash(
          calldataAggregatedProof1To155.l1RollingHashMessageNumber,
          calldataAggregatedProof1To155.l1RollingHash,
        );

        // Set the final state root hash to zero
        finalizationData.shnarfData.finalStateRootHash = HASH_ZERO;

        const finalizeCall = lineaRollup
          .connect(operator)
          .finalizeBlocks(calldataAggregatedProof1To155.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCall, "FinalBlockStateEqualsZeroHash");
      });
    });
  });

  describe("Compressed data finalization with proof", () => {
    beforeEach(async () => {
      await lineaRollup.setLastFinalizedBlock(0);
    });

    it("Should revert if the caller does not have the OPERATOR_ROLE", async () => {
      const finalizationData = await generateFinalizationData();

      const finalizeCall = lineaRollup
        .connect(nonAuthorizedAccount)
        .finalizeBlocks(calldataAggregatedProof1To155.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
      await expectRevertWithReason(finalizeCall, buildAccessErrorMessage(nonAuthorizedAccount, OPERATOR_ROLE));
    });

    it("Should revert if GENERAL_PAUSE_TYPE is enabled", async () => {
      const finalizationData = await generateFinalizationData();

      await lineaRollup.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocks(EMPTY_CALLDATA, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "IsPaused", [GENERAL_PAUSE_TYPE]);
    });

    it("Should revert if FINALIZATION_PAUSE_TYPE is enabled", async () => {
      const finalizationData = await generateFinalizationData();

      await lineaRollup.connect(securityCouncil).pauseByType(FINALIZATION_PAUSE_TYPE);

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocks(EMPTY_CALLDATA, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "IsPaused", [FINALIZATION_PAUSE_TYPE]);
    });

    it("Should revert if the proof is empty", async () => {
      const finalizationData = await generateFinalizationData();

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocks(EMPTY_CALLDATA, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "ProofIsEmpty");
    });

    it("Should revert when finalization parentStateRootHash is different than last finalized state root hash", async () => {
      // Submit 4 sets of compressed data setting the correct shnarf in storage
      const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);

      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: MAX_GAS_LIMIT,
          });
        index++;
      }

      const finalizationData = await generateFinalizationData({
        lastFinalizedTimestamp: DEFAULT_LAST_FINALIZED_TIMESTAMP,
        parentStateRootHash: generateRandomBytes(32),
        aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
      });

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocks(calldataAggregatedProof1To155.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData, {
          gasLimit: MAX_GAS_LIMIT,
        });
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "StartingRootHashDoesNotMatch");
    });

    it("Should successfully finalize with only previously submitted data", async () => {
      // Submit 4 sets of compressed data setting the correct shnarf in storage
      const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);
      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: MAX_GAS_LIMIT,
          });
        index++;
      }

      await expectSuccessfulFinalize(
        lineaRollup,
        operator,
        calldataAggregatedProof1To155,
        index,
        fourthCompressedDataContent.finalStateRootHash,
        generateParentShnarfData,
      );
    });

    it("Should revert when proofType is invalid", async () => {
      const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);
      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: MAX_GAS_LIMIT,
          });
        index++;
      }

      const finalizationData = await generateFinalizationData({
        l1RollingHash: calldataAggregatedProof1To155.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(calldataAggregatedProof1To155.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(calldataAggregatedProof1To155.parentAggregationLastBlockTimestamp),
        endBlockNumber: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
        parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
        finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
        l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
        aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
        shnarfData: generateParentShnarfData(index),
      });

      await lineaRollup.setRollingHash(
        calldataAggregatedProof1To155.l1RollingHashMessageNumber,
        calldataAggregatedProof1To155.l1RollingHash,
      );

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocks(calldataAggregatedProof1To155.aggregatedProof, 99, finalizationData);
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "InvalidProofType");
    });

    it("Should revert when using a proofType index that was removed", async () => {
      const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);
      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: MAX_GAS_LIMIT,
          });
        index++;
      }

      const finalizationData = await generateFinalizationData({
        l1RollingHash: calldataAggregatedProof1To155.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(calldataAggregatedProof1To155.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(calldataAggregatedProof1To155.parentAggregationLastBlockTimestamp),
        endBlockNumber: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
        parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
        finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
        l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
        aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
        shnarfData: generateParentShnarfData(index),
      });

      await lineaRollup.setRollingHash(
        calldataAggregatedProof1To155.l1RollingHashMessageNumber,
        calldataAggregatedProof1To155.l1RollingHash,
      );

      // removing the verifier index
      await lineaRollup.connect(securityCouncil).unsetVerifierAddress(TEST_PUBLIC_VERIFIER_INDEX);

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocks(calldataAggregatedProof1To155.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "InvalidProofType");
    });

    it("Should fail when proof does not match", async () => {
      const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);
      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: MAX_GAS_LIMIT,
          });
        index++;
      }

      const finalizationData = await generateFinalizationData({
        l1RollingHash: calldataAggregatedProof1To155.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(calldataAggregatedProof1To155.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(calldataAggregatedProof1To155.parentAggregationLastBlockTimestamp),
        endBlockNumber: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
        parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
        finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
        l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
        aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
        shnarfData: generateParentShnarfData(index),
      });

      await lineaRollup.setRollingHash(
        calldataAggregatedProof1To155.l1RollingHashMessageNumber,
        calldataAggregatedProof1To155.l1RollingHash,
      );

      // aggregatedProof1To81.aggregatedProof, wrong proof on purpose
      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocks(aggregatedProof1To81.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "InvalidProof");
    });

    it("Should fail if shnarf does not exist when finalizing", async () => {
      const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);
      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: MAX_GAS_LIMIT,
          });
        index++;
      }

      const finalizationData = await generateFinalizationData({
        l1RollingHash: calldataAggregatedProof1To155.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(calldataAggregatedProof1To155.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(calldataAggregatedProof1To155.parentAggregationLastBlockTimestamp),
        endBlockNumber: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
        parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
        finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
        l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
        aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
        shnarfData: generateParentShnarfData(1),
      });

      await lineaRollup.setRollingHash(
        calldataAggregatedProof1To155.l1RollingHashMessageNumber,
        calldataAggregatedProof1To155.l1RollingHash,
      );

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocks(calldataAggregatedProof1To155.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "InvalidProof");
    });

    it("Should successfully finalize 1-81 and then 82-153 in two separate finalizations", async () => {
      const submissionDataBeforeFinalization = generateCallDataSubmissionMultipleProofs(0, 4);
      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForMulitpleIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: MAX_GAS_LIMIT,
          });
        index++;
      }

      await expectSuccessfulFinalize(
        lineaRollup,
        operator,
        aggregatedProof1To81,
        2,
        secondCompressedDataContent.finalStateRootHash,
        generateParentShnarfData,
        true,
      );

      await expectSuccessfulFinalize(
        lineaRollup,
        operator,
        aggregatedProof82To153,
        4,
        fourthMultipleCompressedDataContent.finalStateRootHash,
        generateParentShnarfData,
        true,
        aggregatedProof1To81.l1RollingHash,
        BigInt(aggregatedProof1To81.l1RollingHashMessageNumber),
      );
    });

    it("Should succeed when sending with pure calldata", async () => {
      const submissionDataBeforeFinalization = generateCallDataSubmissionMultipleProofs(0, 4);
      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForMulitpleIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: MAX_GAS_LIMIT,
          });
        index++;
      }
      await lineaRollup.setRollingHash(
        aggregatedProof1To81.l1RollingHashMessageNumber,
        aggregatedProof1To81.l1RollingHash,
      );

      const lineaRollupAddress = await lineaRollup.getAddress();

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();

      const finalizationData = await generateFinalizationData({
        l1RollingHash: aggregatedProof1To81.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(aggregatedProof1To81.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(aggregatedProof1To81.parentAggregationLastBlockTimestamp),
        parentStateRootHash: aggregatedProof1To81.parentStateRootHash,
        finalTimestamp: BigInt(aggregatedProof1To81.finalTimestamp),
        l2MerkleRoots: aggregatedProof1To81.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(aggregatedProof1To81.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: aggregatedProof1To81.l2MessagingBlocksOffsets,
        aggregatedProof: aggregatedProof1To81.aggregatedProof,
        shnarfData: generateParentShnarfData(2, true),
        endBlockNumber: BigInt(aggregatedProof1To81.finalBlockNumber),
      });

      const encodedCall = LineaRollup__factory.createInterface().encodeFunctionData("finalizeBlocks", [
        aggregatedProof1To81.aggregatedProof,
        TEST_PUBLIC_VERIFIER_INDEX,
        finalizationData,
      ]);

      const transaction = {
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: 31337,
        value: 0,
        gasLimit: 5_000_000,
      };

      await expect(operator.sendTransaction(transaction)).to.not.be.reverted;
    });

    it("Should fail when sending with wrong merkle root location", async () => {
      const submissionDataBeforeFinalization = generateCallDataSubmissionMultipleProofs(0, 4);
      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForMulitpleIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: MAX_GAS_LIMIT,
          });
        index++;
      }

      await lineaRollup.setRollingHash(
        aggregatedProof1To81.l1RollingHashMessageNumber,
        aggregatedProof1To81.l1RollingHash,
      );

      const lineaRollupAddress = await lineaRollup.getAddress();

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();

      const finalizationData = await generateFinalizationData({
        l1RollingHash: aggregatedProof1To81.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(aggregatedProof1To81.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(aggregatedProof1To81.parentAggregationLastBlockTimestamp),
        parentStateRootHash: aggregatedProof1To81.parentStateRootHash,
        finalTimestamp: BigInt(aggregatedProof1To81.finalTimestamp),
        l2MerkleRoots: aggregatedProof1To81.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(aggregatedProof1To81.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: aggregatedProof1To81.l2MessagingBlocksOffsets,
        aggregatedProof: aggregatedProof1To81.aggregatedProof,
        shnarfData: generateParentShnarfData(2, true),
        endBlockNumber: BigInt(aggregatedProof1To81.finalBlockNumber),
      });

      finalizationData.lastFinalizedL1RollingHash = HASH_ZERO;
      finalizationData.lastFinalizedL1RollingHashMessageNumber = 0n;

      let encodedCall = LineaRollup__factory.createInterface().encodeFunctionData("finalizeBlocks", [
        aggregatedProof1To81.aggregatedProof,
        TEST_PUBLIC_VERIFIER_INDEX,
        finalizationData,
      ]);

      // Patch the l2MerkleRoots offset pointer (at byte offset 0x220 from finalizationData start)
      // to point to an alternate location, creating a malformed call that should fail with InvalidProof.
      // The finalizationData starts at offset 0x64 (after selector + proof offset + verifierIndex + finalizationData offset).
      // l2MerkleRoots offset is at 0x220 within finalizationData struct.
      // Total offset from start of calldata: 4 (selector) + 0x60 (first 3 words) + 0x220 = 0x284 = 644 bytes = 1288 hex chars + 2 for "0x"
      const patchOffset = 2 + (4 + 0x60 + 0x220) * 2; // 1290
      // Replace the offset value with a contrived pointer (0x0280 -> points to alternate location with value 0x0e)
      encodedCall =
        encodedCall.slice(0, patchOffset) +
        "0000000000000000000000000000000000000000000000000000000000000280" +
        encodedCall.slice(patchOffset + 64);

      // TODO: Remove this after capturing the hardcoded value
      console.log("PATCHED_ENCODED_CALL:", encodedCall);

      const transaction = {
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: 31337,
        value: 0,
        gasLimit: 5_000_000,
      };

      await expectRevertWithCustomError(lineaRollup, operator.sendTransaction(transaction), "InvalidProof");
    });

    it("Should fail to finalize with extra merkle roots", async () => {
      const submissionDataBeforeFinalization = generateCallDataSubmissionMultipleProofs(0, 4);
      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForMulitpleIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: MAX_GAS_LIMIT,
          });
        index++;
      }

      const merkleRoots = aggregatedProof1To81.l2MerkleRoots;
      merkleRoots.push(generateRandomBytes(32));

      const finalizationData = await generateFinalizationData({
        l1RollingHash: aggregatedProof1To81.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(aggregatedProof1To81.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(aggregatedProof1To81.parentAggregationLastBlockTimestamp),
        parentStateRootHash: aggregatedProof1To81.parentStateRootHash,
        finalTimestamp: BigInt(aggregatedProof1To81.finalTimestamp),
        l2MerkleRoots: merkleRoots,
        l2MerkleTreesDepth: BigInt(aggregatedProof1To81.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: aggregatedProof1To81.l2MessagingBlocksOffsets,
        aggregatedProof: aggregatedProof1To81.aggregatedProof,
        shnarfData: generateParentShnarfData(2, true),
      });

      finalizationData.lastFinalizedL1RollingHash = HASH_ZERO;
      finalizationData.lastFinalizedL1RollingHashMessageNumber = 0n;

      await lineaRollup.setRollingHash(
        aggregatedProof1To81.l1RollingHashMessageNumber,
        aggregatedProof1To81.l1RollingHash,
      );

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocks(aggregatedProof1To81.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);

      await expectRevertWithCustomError(lineaRollup, finalizeCall, "InvalidProof");
    });
  });
});
