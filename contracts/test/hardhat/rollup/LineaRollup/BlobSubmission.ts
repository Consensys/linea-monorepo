import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { BaseContract } from "ethers";
import { ethers } from "hardhat";

import blobAggregatedProof1To155 from "../../_testData/compressedDataEip4844/aggregatedProof-1-155.json";
import blobMultipleAggregatedProof1To81 from "../../_testData/compressedDataEip4844/multipleProofs/aggregatedProof-1-81.json";
import blobMultipleAggregatedProof82To153 from "../../_testData/compressedDataEip4844/multipleProofs/aggregatedProof-82-139.json";
import firstCompressedDataContent from "../../_testData/compressedData/blocks-1-46.json";

import { AddressFilter, LineaRollup__factory, TestLineaRollup } from "contracts/typechain-types";
import {
  deployForcedTransactionGatewayFixture,
  deployRevertingVerifier,
  expectFailedCustomErrorFinalize,
  expectSuccessfulFinalize,
  getAccountsFixture,
  getWalletForIndex,
  sendBlobTransaction,
  buildBlobTransaction,
  submitBlobsAndGetReceipt,
} from "../helpers";
import {
  GENERAL_PAUSE_TYPE,
  HASH_ZERO,
  OPERATOR_ROLE,
  TEST_PUBLIC_VERIFIER_INDEX,
  DEFAULT_LAST_FINALIZED_TIMESTAMP,
  STATE_DATA_SUBMISSION_PAUSE_TYPE,
  FORCED_TRANSACTION_FEE,
} from "../../common/constants";
import {
  generateFinalizationData,
  generateRandomBytes,
  generateKeccak256,
  buildAccessErrorMessage,
  expectRevertWithCustomError,
  expectRevertWithReason,
  generateBlobDataSubmission,
  generateBlobParentShnarfData,
  expectEventDirectFromReceiptData,
  proofDataToFinalizationParams,
  expectRevertWhenPaused,
} from "../../common/helpers";
import { AggregatedProofData } from "../../common/types";
import { reinitializeUpgradeableProxy } from "../../common/deployment";

describe("Linea Rollup contract: EIP-4844 Blob submission tests", () => {
  let lineaRollup: TestLineaRollup;
  let revertingVerifier: string;

  let securityCouncil: SignerWithAddress;
  let operator: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;
  let addressFilterAddress: string;
  let addressFilter: AddressFilter;

  const { prevShnarf } = firstCompressedDataContent;

  before(async () => {
    ({ securityCouncil, operator, nonAuthorizedAccount } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ lineaRollup, addressFilter } = await loadFixture(deployForcedTransactionGatewayFixture));

    addressFilterAddress = await addressFilter.getAddress();

    await lineaRollup.setLastFinalizedBlock(0);
    await lineaRollup.setupParentShnarf(prevShnarf);
    await lineaRollup.connect(securityCouncil).setAddressFilter(addressFilterAddress);
  });

  it("Should successfully submit blobs", async () => {
    const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

    const receipt = await submitBlobsAndGetReceipt({
      lineaRollup,
      blobSubmission: blobDataSubmission,
      compressedBlobs,
      parentShnarf,
      finalShnarf,
    });

    expect(receipt).is.not.null;

    const expectedEventArgs = [
      parentShnarf,
      finalShnarf,
      blobDataSubmission[blobDataSubmission.length - 1].finalStateRootHash,
    ];

    expectEventDirectFromReceiptData(lineaRollup as BaseContract, receipt!, "DataSubmittedV3", expectedEventArgs);

    const blobShnarfExists = await lineaRollup.blobShnarfExists(finalShnarf);
    expect(blobShnarfExists).to.equal(1n);
  });

  it("Fails the blob submission when the parent shnarf is missing", async () => {
    const lineaRollupAddress = await lineaRollup.getAddress();
    const { blobDataSubmission, compressedBlobs } = generateBlobDataSubmission(0, 1);
    const nonExistingParentShnarf = generateRandomBytes(32);

    const wrongExpectedShnarf = generateKeccak256(
      ["bytes32", "bytes32", "bytes32", "bytes32", "bytes32"],
      [HASH_ZERO, HASH_ZERO, blobDataSubmission[0].finalStateRootHash, HASH_ZERO, HASH_ZERO],
    );

    const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
      blobDataSubmission,
      nonExistingParentShnarf,
      wrongExpectedShnarf,
    ]);

    const transaction = await buildBlobTransaction({
      lineaRollupAddress,
      encodedCall,
      compressedBlobs,
    });

    const signedTx = await getWalletForIndex(2).signTransaction(transaction);

    await expectRevertWithCustomError(
      lineaRollup,
      ethers.provider.broadcastTransaction(signedTx),
      "ParentShnarfNotSubmitted",
      [nonExistingParentShnarf],
    );
  });

  it("Fails when the blob submission data is missing", async () => {
    const lineaRollupAddress = await lineaRollup.getAddress();
    const { compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

    const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [[], parentShnarf, finalShnarf]);

    const transaction = await buildBlobTransaction({
      lineaRollupAddress,
      encodedCall,
      compressedBlobs,
    });

    const signedTx = await getWalletForIndex(2).signTransaction(transaction);

    await expectRevertWithCustomError(
      lineaRollup,
      ethers.provider.broadcastTransaction(signedTx),
      "BlobSubmissionDataIsMissing",
    );
  });

  it("Should revert if the caller does not have the OPERATOR_ROLE", async () => {
    const { blobDataSubmission, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

    await expectRevertWithReason(
      lineaRollup.connect(nonAuthorizedAccount).submitBlobs(blobDataSubmission, parentShnarf, finalShnarf),
      buildAccessErrorMessage(nonAuthorizedAccount, OPERATOR_ROLE),
    );
  });

  // Parameterized pause type tests for blob submission
  const blobSubmissionPauseTypes = [
    { pauseType: GENERAL_PAUSE_TYPE, name: "GENERAL_PAUSE_TYPE" },
    { pauseType: STATE_DATA_SUBMISSION_PAUSE_TYPE, name: "STATE_DATA_SUBMISSION_PAUSE_TYPE" },
  ];

  blobSubmissionPauseTypes.forEach(({ pauseType, name }) => {
    it(`Should revert if ${name} is enabled`, async () => {
      const { blobDataSubmission, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

      await lineaRollup.connect(securityCouncil).pauseByType(pauseType);

      await expectRevertWhenPaused(
        lineaRollup,
        lineaRollup.connect(operator).submitBlobs(blobDataSubmission, parentShnarf, finalShnarf),
        pauseType,
      );
    });
  });

  it("Should revert if the blob data is empty at any index", async () => {
    const lineaRollupAddress = await lineaRollup.getAddress();
    const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 2);

    const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
      blobDataSubmission,
      parentShnarf,
      finalShnarf,
    ]);

    // Pass only the first blob to simulate empty blob at index 1
    const transaction = await buildBlobTransaction({
      lineaRollupAddress,
      encodedCall,
      compressedBlobs: [compressedBlobs[0]],
    });

    const signedTx = await getWalletForIndex(2).signTransaction(transaction);

    await expectRevertWithCustomError(
      lineaRollup,
      ethers.provider.broadcastTransaction(signedTx),
      "EmptyBlobDataAtIndex",
      [1n],
    );
  });

  it("Should fail if the final state root hash is empty", async () => {
    const lineaRollupAddress = await lineaRollup.getAddress();
    const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

    blobDataSubmission[0].finalStateRootHash = HASH_ZERO;

    const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
      blobDataSubmission,
      parentShnarf,
      finalShnarf,
    ]);

    const transaction = await buildBlobTransaction({
      lineaRollupAddress,
      encodedCall,
      compressedBlobs,
    });

    const signedTx = await getWalletForIndex(2).signTransaction(transaction);

    // TODO: Make the failure shnarf dynamic and computed
    await expectRevertWithCustomError(lineaRollup, ethers.provider.broadcastTransaction(signedTx), "FinalShnarfWrong", [
      finalShnarf,
      "0x22f8fb954df8328627fe9c48b60192f4d970a92891417aaadea39300ca244d36",
    ]);
  });

  it("Should revert when snarkHash is zero hash", async () => {
    const lineaRollupAddress = await lineaRollup.getAddress();
    const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

    // Set the snarkHash to a random value for a specific index
    blobDataSubmission[0].snarkHash = generateRandomBytes(32);

    const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
      blobDataSubmission,
      parentShnarf,
      finalShnarf,
    ]);

    const transaction = await buildBlobTransaction({
      lineaRollupAddress,
      encodedCall,
      compressedBlobs,
    });

    const signedTx = await getWalletForIndex(2).signTransaction(transaction);

    await expectRevertWithCustomError(
      lineaRollup,
      ethers.provider.broadcastTransaction(signedTx),
      "PointEvaluationFailed",
    );
  });

  it("Should revert if the final shnarf is wrong", async () => {
    const lineaRollupAddress = await lineaRollup.getAddress();
    const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 2);
    const badFinalShnarf = generateRandomBytes(32);

    const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
      blobDataSubmission,
      parentShnarf,
      badFinalShnarf,
    ]);

    const transaction = await buildBlobTransaction({
      lineaRollupAddress,
      encodedCall,
      compressedBlobs,
    });

    const signedTx = await getWalletForIndex(2).signTransaction(transaction);

    await expectRevertWithCustomError(lineaRollup, ethers.provider.broadcastTransaction(signedTx), "FinalShnarfWrong", [
      badFinalShnarf,
      finalShnarf,
    ]);
  });

  it("Should revert if the data has already been submitted", async () => {
    await sendBlobTransaction(lineaRollup, 0, 1);

    const lineaRollupAddress = await lineaRollup.getAddress();
    const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

    // Try to submit the same blob data again
    const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
      blobDataSubmission,
      parentShnarf,
      finalShnarf,
    ]);

    const transaction = await buildBlobTransaction({
      lineaRollupAddress,
      encodedCall,
      compressedBlobs,
    });

    const signedTx = await getWalletForIndex(2).signTransaction(transaction);

    await expectRevertWithCustomError(
      lineaRollup,
      ethers.provider.broadcastTransaction(signedTx),
      "ShnarfAlreadySubmitted",
      [finalShnarf],
    );
  });

  it("Should revert with PointEvaluationFailed when point evaluation fails", async () => {
    const lineaRollupAddress = await lineaRollup.getAddress();
    const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

    // Modify the kzgProof to an invalid value to trigger the PointEvaluationFailed revert
    blobDataSubmission[0].kzgProof = HASH_ZERO;

    const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
      blobDataSubmission,
      parentShnarf,
      finalShnarf,
    ]);

    const transaction = await buildBlobTransaction({
      lineaRollupAddress,
      encodedCall,
      compressedBlobs,
    });

    const signedTx = await getWalletForIndex(2).signTransaction(transaction);

    await expectRevertWithCustomError(
      lineaRollup,
      ethers.provider.broadcastTransaction(signedTx),
      "PointEvaluationFailed",
    );
  });

  it("Should submit 2 blobs, then submit another 2 blobs and finalize", async () => {
    // we need the address filter to be set
    await reinitializeUpgradeableProxy(lineaRollup, LineaRollup__factory.abi, "reinitializeLineaRollupV9", [
      FORCED_TRANSACTION_FEE,
      addressFilterAddress,
    ]);

    // validating address filtering checking by marking the security council as filtered
    await addressFilter.connect(securityCouncil).setFilteredStatus([securityCouncil.getAddress()], true);

    // Submit 2 blobs
    await sendBlobTransaction(lineaRollup, 0, 2);
    // Submit another 2 blobs
    await sendBlobTransaction(lineaRollup, 2, 4);
    // Finalize 4 blobs
    await expectSuccessfulFinalize({
      context: {
        lineaRollup,
        operator,
      },
      proofConfig: {
        proofData: blobAggregatedProof1To155,
        blobParentShnarfIndex: 4,
        shnarfDataGenerator: generateBlobParentShnarfData,
        isMultiple: false,
      },
    });
  });

  it("Should revert if the address filter is set and the address is not marked as filtered", async () => {
    const filteredAddress = await securityCouncil.getAddress();

    // we need the address filter to be set
    await reinitializeUpgradeableProxy(lineaRollup, LineaRollup__factory.abi, "reinitializeLineaRollupV9", [
      FORCED_TRANSACTION_FEE,
      addressFilterAddress,
    ]);

    // Submit 2 blobs
    await sendBlobTransaction(lineaRollup, 0, 2);
    // Submit another 2 blobs
    await sendBlobTransaction(lineaRollup, 2, 4);
    // Finalize 4 blobs
    await expectFailedCustomErrorFinalize({
      context: {
        lineaRollup,
        operator,
      },
      proofConfig: {
        proofData: blobAggregatedProof1To155,
        blobParentShnarfIndex: 4,
        shnarfDataGenerator: generateBlobParentShnarfData,
        isMultiple: false,
      },
      overrides: {
        filteredAddresses: [filteredAddress],
      },
      expectedError: {
        name: "AddressIsNotFiltered",
        args: [filteredAddress],
      },
    });
  });

  it("Should revert if there is less data than blobs", async () => {
    const lineaRollupAddress = await lineaRollup.getAddress();

    const {
      blobDataSubmission: blobSubmission,
      compressedBlobs,
      parentShnarf,
      finalShnarf,
    } = generateBlobDataSubmission(0, 2, true);

    const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
      [blobSubmission[0]],
      parentShnarf,
      finalShnarf,
    ]);

    const transaction = await buildBlobTransaction({
      lineaRollupAddress,
      encodedCall,
      compressedBlobs,
    });

    const signedTx = await getWalletForIndex(2).signTransaction(transaction);

    await expectRevertWithCustomError(
      lineaRollup,
      ethers.provider.broadcastTransaction(signedTx),
      "BlobSubmissionDataEmpty",
      [1],
    );
  });

  it("Should fail to finalize with not enough gas for the rollup (pre-verifier)", async () => {
    // Submit 2 blobs
    await sendBlobTransaction(lineaRollup, 0, 2);
    // Submit another 2 blobs
    await sendBlobTransaction(lineaRollup, 2, 4);

    const proofData = blobAggregatedProof1To155 as AggregatedProofData;
    const finalizationData = await generateFinalizationData({
      ...proofDataToFinalizationParams({
        proofData,
        shnarfDataGenerator: generateBlobParentShnarfData,
        blobParentShnarfIndex: 4,
        isMultiple: false,
      }),
      lastFinalizedL1RollingHash: HASH_ZERO,
      lastFinalizedL1RollingHashMessageNumber: 0n,
    });

    await lineaRollup.setRollingHash(proofData.l1RollingHashMessageNumber, proofData.l1RollingHash);

    const finalizeCompressedCall = lineaRollup
      .connect(operator)
      .finalizeBlocks(proofData.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData, {
        gasLimit: 80_000,
      });

    // there is no reason
    await expect(finalizeCompressedCall).to.be.reverted;
  });

  it("Should fail to finalize with not enough gas to verify", async () => {
    // Submit 2 blobs
    await sendBlobTransaction(lineaRollup, 0, 2);
    // Submit another 2 blobs
    await sendBlobTransaction(lineaRollup, 2, 4);

    const proofData = blobAggregatedProof1To155 as AggregatedProofData;
    const finalizationData = await generateFinalizationData({
      ...proofDataToFinalizationParams({
        proofData,
        shnarfDataGenerator: generateBlobParentShnarfData,
        blobParentShnarfIndex: 4,
        isMultiple: false,
      }),
      lastFinalizedL1RollingHash: HASH_ZERO,
      lastFinalizedL1RollingHashMessageNumber: 0n,
    });

    await lineaRollup.setRollingHash(proofData.l1RollingHashMessageNumber, proofData.l1RollingHash);

    const finalizeCompressedCall = lineaRollup
      .connect(operator)
      .finalizeBlocks(proofData.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData, {
        gasLimit: 400000,
      });

    await expectRevertWithCustomError(
      lineaRollup,
      finalizeCompressedCall,
      "InvalidProofOrProofVerificationRanOutOfGas",
      ["error pairing"],
    );
  });

  const testCases = [
    { revertScenario: 0n, title: "Should fail to finalize via EMPTY_REVERT scenario with 'Unknown'" },
    { revertScenario: 1n, title: "Should fail to finalize via GAS_GUZZLE scenario with 'Unknown'" },
  ];

  testCases.forEach(({ revertScenario, title }) => {
    it(title, async () => {
      revertingVerifier = await deployRevertingVerifier(revertScenario);
      await lineaRollup.connect(securityCouncil).setVerifierAddress(revertingVerifier, 0);

      // Submit 2 blobs
      await sendBlobTransaction(lineaRollup, 0, 2);
      // Submit another 2 blobs
      await sendBlobTransaction(lineaRollup, 2, 4);

      const proofData = blobAggregatedProof1To155 as AggregatedProofData;
      const finalizationData = await generateFinalizationData({
        ...proofDataToFinalizationParams({
          proofData,
          shnarfDataGenerator: generateBlobParentShnarfData,
          blobParentShnarfIndex: 4,
          isMultiple: false,
        }),
        lastFinalizedL1RollingHash: HASH_ZERO,
        lastFinalizedL1RollingHashMessageNumber: 0n,
      });

      await lineaRollup.setRollingHash(proofData.l1RollingHashMessageNumber, proofData.l1RollingHash);

      const finalizeCompressedCall = lineaRollup
        .connect(operator)
        .finalizeBlocks(proofData.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData, {
          gasLimit: 400000,
        });

      await expectRevertWithCustomError(
        lineaRollup,
        finalizeCompressedCall,
        "InvalidProofOrProofVerificationRanOutOfGas",
        ["Unknown"],
      );
    });
  });

  it("Should fail to finalize if there are missing forced transactions", async () => {
    // Submit 2 blobs
    await sendBlobTransaction(lineaRollup, 0, 2, true);
    // Submit another 2 blobs
    await sendBlobTransaction(lineaRollup, 2, 4, true);

    const proofData = blobAggregatedProof1To155 as AggregatedProofData;
    await lineaRollup.setForcedTransactionBlockNumber(BigInt(proofData.finalBlockNumber));

    const expectedErrorTransactionNumber = 1; // first transaction

    const finalizationData = await generateFinalizationData({
      ...proofDataToFinalizationParams({
        proofData,
        shnarfDataGenerator: generateBlobParentShnarfData,
        blobParentShnarfIndex: 4,
        isMultiple: false,
      }),
      parentStateRootHash: HASH_ZERO, // Manipulate for bypass
      lastFinalizedL1RollingHash: HASH_ZERO,
      lastFinalizedL1RollingHashMessageNumber: 0n,
    });

    await lineaRollup.setRollingHash(proofData.l1RollingHashMessageNumber, proofData.l1RollingHash);

    await lineaRollup.setLastFinalizedBlock(10_000_000);

    await expectRevertWithCustomError(
      lineaRollup,
      lineaRollup
        .connect(operator)
        .finalizeBlocks(proofData.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData),
      "FinalizationDataMissingForcedTransaction",
      [expectedErrorTransactionNumber],
    );
  });

  it("Should successfully submit 2 blobs twice then finalize in two separate finalizations", async () => {
    // Submit 2 blobs
    await sendBlobTransaction(lineaRollup, 0, 2, true);
    // Submit another 2 blobs
    await sendBlobTransaction(lineaRollup, 2, 4, true);

    await lineaRollup.setForcedTransactionRollingHash(
      blobMultipleAggregatedProof1To81.finalFtxNumber,
      blobMultipleAggregatedProof1To81.finalFtxRollingHash,
    );
    await lineaRollup.setForcedTransactionRollingHash(
      blobMultipleAggregatedProof82To153.finalFtxNumber,
      blobMultipleAggregatedProof82To153.finalFtxRollingHash,
    );

    for (const filteredAddress of blobMultipleAggregatedProof1To81.filteredAddresses) {
      await addressFilter.connect(securityCouncil).setFilteredStatus([filteredAddress], true);
    }

    await expectSuccessfulFinalize({
      context: {
        lineaRollup,
        operator,
      },
      proofConfig: {
        proofData: blobMultipleAggregatedProof1To81,
        blobParentShnarfIndex: 2,
        shnarfDataGenerator: generateBlobParentShnarfData,
        isMultiple: true,
      },
    });

    for (const filteredAddress of blobMultipleAggregatedProof82To153.filteredAddresses) {
      await addressFilter.connect(securityCouncil).setFilteredStatus([filteredAddress], true);
    }

    // Finalize second 2 blobs
    await expectSuccessfulFinalize({
      context: {
        lineaRollup,
        operator,
      },
      proofConfig: {
        proofData: blobMultipleAggregatedProof82To153,
        blobParentShnarfIndex: 4,
        shnarfDataGenerator: generateBlobParentShnarfData,
        isMultiple: true,
      },
    });
  });

  it("Should successfully submit 2 blobs twice then finalize in two separate finalizations using 3 and then 6 finalizationState fields", async () => {
    // Explicitly use the 3 fields to simulate an existing finalization
    await lineaRollup.setLastFinalizedStateV6(0, HASH_ZERO, DEFAULT_LAST_FINALIZED_TIMESTAMP);

    // Submit 2 blobs
    await sendBlobTransaction(lineaRollup, 0, 2, true);
    // Submit another 2 blobs
    await sendBlobTransaction(lineaRollup, 2, 4, true);

    await lineaRollup.setForcedTransactionRollingHash(
      blobMultipleAggregatedProof1To81.finalFtxNumber,
      blobMultipleAggregatedProof1To81.finalFtxRollingHash,
    );

    for (const filteredAddress of blobMultipleAggregatedProof1To81.filteredAddresses) {
      await addressFilter.connect(securityCouncil).setFilteredStatus([filteredAddress], true);
    }

    // Finalize first 2 blobs
    await expectSuccessfulFinalize({
      context: {
        lineaRollup,
        operator,
      },
      proofConfig: {
        proofData: blobMultipleAggregatedProof1To81,
        blobParentShnarfIndex: 2,
        shnarfDataGenerator: generateBlobParentShnarfData,
        isMultiple: true,
      },
    });

    await lineaRollup.setForcedTransactionRollingHash(
      blobMultipleAggregatedProof82To153.finalFtxNumber,
      blobMultipleAggregatedProof82To153.finalFtxRollingHash,
    );

    for (const filteredAddress of blobMultipleAggregatedProof82To153.filteredAddresses) {
      await addressFilter.connect(securityCouncil).setFilteredStatus([filteredAddress], true);
    }

    // Finalize second 2 blobs
    await expectSuccessfulFinalize({
      context: {
        lineaRollup,
        operator,
      },
      proofConfig: {
        proofData: blobMultipleAggregatedProof82To153,
        blobParentShnarfIndex: 4,
        shnarfDataGenerator: generateBlobParentShnarfData,
        isMultiple: true,
      },
    });
  });

  it("Should fail to prove if last finalized is higher than proving range", async () => {
    // Submit 2 blobs
    await sendBlobTransaction(lineaRollup, 0, 2, false);
    // Submit another 2 blobs
    await sendBlobTransaction(lineaRollup, 2, 4, false);

    await lineaRollup.setLastFinalizedBlock(10_000_000);

    await expectFailedCustomErrorFinalize({
      context: {
        lineaRollup,
        operator,
      },
      proofConfig: {
        proofData: blobAggregatedProof1To155,
        blobParentShnarfIndex: 4,
        shnarfDataGenerator: generateBlobParentShnarfData,
        isMultiple: false,
      },
      overrides: {
        parentStateRootHash: HASH_ZERO,
      },
      expectedError: {
        name: "InvalidProof",
      },
    });
  });
});
