import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture, time as networkTime } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { config, ethers, upgrades } from "hardhat";
import { HardhatNetworkHDAccountsConfig } from "hardhat/types";
import { HDNodeWallet, Transaction, Wallet, ZeroAddress } from "ethers";
import { TestLineaRollup } from "../typechain-types";
import calldataAggregatedProof1To155 from "./testData/compressedData/aggregatedProof-1-155.json";
import blobAggregatedProof1To155 from "./testData/compressedDataEip4844/aggregatedProof-1-155.json";
import blobMultipleAggregatedProof1To81 from "./testData/compressedDataEip4844/multipleProofs/aggregatedProof-1-81.json";
import blobMultipleAggregatedProof82To153 from "./testData/compressedDataEip4844/multipleProofs/aggregatedProof-82-153.json";
import firstCompressedDataContent from "./testData/compressedData/blocks-1-46.json";
import secondCompressedDataContent from "./testData/compressedData/blocks-47-81.json";
import fourthCompressedDataContent from "./testData/compressedData/blocks-115-155.json";
import fourthMultipleBlobDataContent from "./testData/compressedDataEip4844/multipleProofs/blocks-120-153.json";
import fourthMultipleCompressedDataContent from "./testData/compressedData/multipleProofs/blocks-120-153.json";
import {
  ADDRESS_ZERO,
  GENERAL_PAUSE_TYPE,
  HASH_WITHOUT_ZERO_FIRST_BYTE,
  HASH_ZERO,
  INITIAL_MIGRATION_BLOCK,
  INITIAL_WITHDRAW_LIMIT,
  ONE_DAY_IN_SECONDS,
  OPERATOR_ROLE,
  TEST_PUBLIC_VERIFIER_INDEX,
  VERIFIER_SETTER_ROLE,
  VERIFIER_UNSETTER_ROLE,
  GENESIS_L2_TIMESTAMP,
  EMPTY_CALLDATA,
  INITIALIZED_ALREADY_MESSAGE,
  FINALIZE_WITHOUT_PROOF_ROLE,
  CALLDATA_SUBMISSION_PAUSE_TYPE,
  BLOB_SUBMISSION_PAUSE_TYPE,
  FINALIZATION_PAUSE_TYPE,
  PAUSE_ALL_ROLE,
  unpauseTypeRoles,
  pauseTypeRoles,
  DEFAULT_ADMIN_ROLE,
  UNPAUSE_ALL_ROLE,
  PAUSE_L2_BLOB_SUBMISSION_ROLE,
  UNPAUSE_L2_BLOB_SUBMISSION_ROLE,
  PAUSE_FINALIZE_WITHPROOF_ROLE,
  UNPAUSE_FINALIZE_WITHPROOF_ROLE,
  LINEA_ROLLUP_INITIALIZE_SIGNATURE,
} from "./utils/constants";
import { deployUpgradableFromFactory } from "./utils/deployment";
import {
  calculateRollingHash,
  encodeData,
  generateFinalizationData,
  generateRandomBytes,
  generateCallDataSubmission,
  generateCallDataSubmissionMultipleProofs,
  generateParentSubmissionDataForIndex,
  generateKeccak256,
  expectEvent,
  buildAccessErrorMessage,
  expectRevertWithCustomError,
  expectRevertWithReason,
  generateParentAndExpectedShnarfForIndex,
  generateParentAndExpectedShnarfForMulitpleIndex,
  generateParentShnarfData,
  generateBlobDataSubmission,
  generateBlobParentShnarfData,
  ShnarfDataGenerator,
} from "./utils/helpers";
import { CalldataSubmissionData } from "./utils/types";
import aggregatedProof1To81 from "./testData/compressedData/multipleProofs/aggregatedProof-1-81.json";
import aggregatedProof82To153 from "./testData/compressedData/multipleProofs/aggregatedProof-82-153.json";
import * as kzg from "c-kzg";

kzg.loadTrustedSetup(`${__dirname}/testData/trusted_setup.txt`);

describe("Linea Rollup contract", () => {
  let lineaRollup: TestLineaRollup;
  let revertingVerifier: string;

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let admin: SignerWithAddress;
  let verifier: string;
  let securityCouncil: SignerWithAddress;
  let operator: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;

  const multiCallAddress = "0xcA11bde05977b3631167028862bE2a173976CA11";

  const { compressedData, prevShnarf, expectedShnarf, expectedX, expectedY, parentDataHash, parentStateRootHash } =
    firstCompressedDataContent;
  const { expectedShnarf: secondExpectedShnarf } = secondCompressedDataContent;

  async function deployRevertingVerifier(scenario: bigint) {
    const RevertingVerifierFactory = await ethers.getContractFactory("RevertingVerifier");
    const verifier = await RevertingVerifierFactory.deploy(scenario);
    await verifier.waitForDeployment();
    revertingVerifier = await verifier.getAddress();
  }

  async function deployLineaRollupFixture() {
    const PlonkVerifierFactory = await ethers.getContractFactory("TestPlonkVerifierForDataAggregation");
    const plonkVerifier = await PlonkVerifierFactory.deploy();
    await plonkVerifier.waitForDeployment();

    verifier = await plonkVerifier.getAddress();

    const roleAddresses = [
      { addressWithRole: securityCouncil.address, role: DEFAULT_ADMIN_ROLE },
      { addressWithRole: securityCouncil.address, role: VERIFIER_SETTER_ROLE },
      { addressWithRole: securityCouncil.address, role: VERIFIER_UNSETTER_ROLE },
      { addressWithRole: securityCouncil.address, role: PAUSE_ALL_ROLE },
      { addressWithRole: securityCouncil.address, role: UNPAUSE_ALL_ROLE },
      { addressWithRole: securityCouncil.address, role: PAUSE_L2_BLOB_SUBMISSION_ROLE },
      { addressWithRole: securityCouncil.address, role: UNPAUSE_L2_BLOB_SUBMISSION_ROLE },
      { addressWithRole: securityCouncil.address, role: PAUSE_FINALIZE_WITHPROOF_ROLE },
      { addressWithRole: securityCouncil.address, role: UNPAUSE_FINALIZE_WITHPROOF_ROLE },
      { addressWithRole: securityCouncil.address, role: FINALIZE_WITHOUT_PROOF_ROLE },
      { addressWithRole: operator.address, role: OPERATOR_ROLE },
    ];

    const initializationData = {
      initialStateRootHash: parentStateRootHash,
      initialL2BlockNumber: 0,
      genesisTimestamp: 1683325137n,
      defaultVerifier: verifier,
      rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
      rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
      roleAddresses: roleAddresses,
      pauseTypeRoles: pauseTypeRoles,
      unpauseTypeRoles: unpauseTypeRoles,
      gatewayOperator: multiCallAddress,
    };

    const lineaRollup = (await deployUpgradableFromFactory("TestLineaRollup", [initializationData], {
      initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
      unsafeAllow: ["constructor"],
    })) as unknown as TestLineaRollup;

    return lineaRollup;
  }

  const getWalletForIndex = (index: number) => {
    const accounts = config.networks.hardhat.accounts as HardhatNetworkHDAccountsConfig;
    const signer = HDNodeWallet.fromPhrase(accounts.mnemonic, "", `m/44'/60'/0'/0/${index}`);
    return new Wallet(signer.privateKey, ethers.provider);
  };

  before(async () => {
    [admin, securityCouncil, operator, nonAuthorizedAccount] = await ethers.getSigners();
  });

  beforeEach(async () => {
    lineaRollup = await loadFixture(deployLineaRollupFixture);
  });

  describe("Fallback/Receive tests", () => {
    const sendEthToContract = async (data: string) => {
      return admin.sendTransaction({ to: await lineaRollup.getAddress(), value: INITIAL_WITHDRAW_LIMIT, data });
    };

    it("Should fail to send eth to the lineaRollup contract through the fallback", async () => {
      await expect(sendEthToContract(EMPTY_CALLDATA)).to.be.reverted;
    });

    it("Should fail to send eth to the lineaRollup contract through the receive function", async () => {
      await expect(sendEthToContract("0x1234")).to.be.reverted;
    });
  });

  describe("Initialisation", () => {
    it("Should revert if verifier address is zero address ", async () => {
      const initializationData = {
        initialStateRootHash: parentStateRootHash,
        initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
        genesisTimestamp: GENESIS_L2_TIMESTAMP,
        defaultVerifier: ADDRESS_ZERO,
        rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
        rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
        roleAddresses: [
          { addressWithRole: securityCouncil.address, role: DEFAULT_ADMIN_ROLE },
          { addressWithRole: securityCouncil.address, role: VERIFIER_SETTER_ROLE },
        ],
        pauseTypeRoles: pauseTypeRoles,
        unpauseTypeRoles: unpauseTypeRoles,
        gatewayOperator: multiCallAddress,
      };

      const deployCall = deployUpgradableFromFactory("contracts/LineaRollup.sol:LineaRollup", [initializationData], {
        initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
        unsafeAllow: ["constructor"],
      });

      await expectRevertWithCustomError(lineaRollup, deployCall, "ZeroAddressNotAllowed");
    });

    it("Should revert if an operator address is zero address ", async () => {
      const initializationData = {
        initialStateRootHash: parentStateRootHash,
        initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
        genesisTimestamp: GENESIS_L2_TIMESTAMP,
        defaultVerifier: verifier,
        rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
        rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
        roleAddresses: [
          { addressWithRole: securityCouncil.address, role: DEFAULT_ADMIN_ROLE },
          { addressWithRole: securityCouncil.address, role: VERIFIER_SETTER_ROLE },
          { addressWithRole: ADDRESS_ZERO, role: OPERATOR_ROLE },
        ],
        pauseTypeRoles: pauseTypeRoles,
        unpauseTypeRoles: unpauseTypeRoles,
        gatewayOperator: multiCallAddress,
      };

      const deployCall = deployUpgradableFromFactory("TestLineaRollup", [initializationData], {
        initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
        unsafeAllow: ["constructor"],
      });

      await expectRevertWithCustomError(lineaRollup, deployCall, "ZeroAddressNotAllowed");
    });

    it("Should store verifier address in storage ", async () => {
      lineaRollup = await loadFixture(deployLineaRollupFixture);
      expect(await lineaRollup.verifiers(0)).to.be.equal(verifier);
    });

    it("Should assign the OPERATOR_ROLE to operator addresses", async () => {
      lineaRollup = await loadFixture(deployLineaRollupFixture);
      expect(await lineaRollup.hasRole(OPERATOR_ROLE, operator.address)).to.be.true;
    });

    it("Should assign the VERIFIER_SETTER_ROLE to securityCouncil addresses", async () => {
      lineaRollup = await loadFixture(deployLineaRollupFixture);
      expect(await lineaRollup.hasRole(VERIFIER_SETTER_ROLE, securityCouncil.address)).to.be.true;
    });

    it("Should assign the VERIFIER_UNSETTER_ROLE to securityCouncil addresses", async () => {
      lineaRollup = await loadFixture(deployLineaRollupFixture);
      expect(await lineaRollup.hasRole(VERIFIER_UNSETTER_ROLE, securityCouncil.address)).to.be.true;
    });

    it("Should store the startingRootHash in storage for the first block number", async () => {
      const initializationData = {
        initialStateRootHash: parentStateRootHash,
        initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
        genesisTimestamp: GENESIS_L2_TIMESTAMP,
        defaultVerifier: verifier,
        rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
        rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
        roleAddresses: [
          { addressWithRole: securityCouncil.address, role: DEFAULT_ADMIN_ROLE },
          { addressWithRole: securityCouncil.address, role: VERIFIER_SETTER_ROLE },
        ],
        pauseTypeRoles: pauseTypeRoles,
        unpauseTypeRoles: unpauseTypeRoles,
        gatewayOperator: multiCallAddress,
      };

      const lineaRollup = await deployUpgradableFromFactory(
        "contracts/LineaRollup.sol:LineaRollup",
        [initializationData],
        {
          initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
          unsafeAllow: ["constructor"],
        },
      );

      expect(await lineaRollup.stateRootHashes(INITIAL_MIGRATION_BLOCK)).to.be.equal(parentStateRootHash);
    });

    it("Should assign the VERIFIER_SETTER_ROLE to both SecurityCouncil and Operator", async () => {
      const initializationData = {
        initialStateRootHash: parentStateRootHash,
        initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
        genesisTimestamp: GENESIS_L2_TIMESTAMP,
        defaultVerifier: verifier,
        rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
        rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
        roleAddresses: [
          { addressWithRole: securityCouncil.address, role: DEFAULT_ADMIN_ROLE },
          { addressWithRole: securityCouncil.address, role: VERIFIER_SETTER_ROLE },
          { addressWithRole: operator.address, role: VERIFIER_SETTER_ROLE },
        ],
        pauseTypeRoles: pauseTypeRoles,
        unpauseTypeRoles: unpauseTypeRoles,
        gatewayOperator: multiCallAddress,
      };

      const lineaRollup = await deployUpgradableFromFactory(
        "contracts/LineaRollup.sol:LineaRollup",
        [initializationData],
        {
          initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
          unsafeAllow: ["constructor"],
        },
      );

      expect(await lineaRollup.hasRole(VERIFIER_SETTER_ROLE, securityCouncil.address)).to.be.true;
      expect(await lineaRollup.hasRole(VERIFIER_SETTER_ROLE, operator.address)).to.be.true;
    });

    it("Should revert if the initialize function is called a second time", async () => {
      lineaRollup = await loadFixture(deployLineaRollupFixture);
      const initializeCall = lineaRollup.initialize({
        initialStateRootHash: parentStateRootHash,
        initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
        genesisTimestamp: GENESIS_L2_TIMESTAMP,
        defaultVerifier: verifier,
        rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
        rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
        roleAddresses: [
          { addressWithRole: securityCouncil.address, role: DEFAULT_ADMIN_ROLE },
          { addressWithRole: securityCouncil.address, role: VERIFIER_SETTER_ROLE },
        ],
        pauseTypeRoles: pauseTypeRoles,
        unpauseTypeRoles: unpauseTypeRoles,
        gatewayOperator: multiCallAddress,
      });

      await expectRevertWithReason(initializeCall, INITIALIZED_ALREADY_MESSAGE);
    });
  });

  describe("Change verifier address", () => {
    it("Should revert if the caller has not the VERIFIER_SETTER_ROLE", async () => {
      const setVerifierCall = lineaRollup.connect(nonAuthorizedAccount).setVerifierAddress(verifier, 2);

      await expectRevertWithReason(
        setVerifierCall,
        buildAccessErrorMessage(nonAuthorizedAccount, VERIFIER_SETTER_ROLE),
      );
    });

    it("Should revert if the address being set is the zero address", async () => {
      await lineaRollup.connect(securityCouncil).grantRole(VERIFIER_SETTER_ROLE, securityCouncil.address);

      const setVerifierCall = lineaRollup.connect(securityCouncil).setVerifierAddress(ADDRESS_ZERO, 2);
      await expectRevertWithCustomError(lineaRollup, setVerifierCall, "ZeroAddressNotAllowed");
    });

    it("Should set the new verifier address", async () => {
      await lineaRollup.connect(securityCouncil).grantRole(VERIFIER_SETTER_ROLE, securityCouncil.address);

      await lineaRollup.connect(securityCouncil).setVerifierAddress(verifier, 2);
      expect(await lineaRollup.verifiers(2)).to.be.equal(verifier);
    });

    it("Should remove verifier address in storage ", async () => {
      lineaRollup = await loadFixture(deployLineaRollupFixture);
      await lineaRollup.connect(securityCouncil).unsetVerifierAddress(0);

      expect(await lineaRollup.verifiers(0)).to.be.equal(ADDRESS_ZERO);
    });

    it("Should revert when removing verifier address if the caller has not the VERIFIER_UNSETTER_ROLE ", async () => {
      lineaRollup = await loadFixture(deployLineaRollupFixture);

      await expect(lineaRollup.connect(nonAuthorizedAccount).unsetVerifierAddress(0)).to.be.revertedWith(
        buildAccessErrorMessage(nonAuthorizedAccount, VERIFIER_UNSETTER_ROLE),
      );
    });

    it("Should emit the correct event", async () => {
      await lineaRollup.connect(securityCouncil).grantRole(VERIFIER_SETTER_ROLE, securityCouncil.address);

      const oldVerifierAddress = await lineaRollup.verifiers(2);

      const setVerifierCall = lineaRollup.connect(securityCouncil).setVerifierAddress(verifier, 2);
      let expectedArgs = [verifier, 2, securityCouncil.address, oldVerifierAddress];

      await expectEvent(lineaRollup, setVerifierCall, "VerifierAddressChanged", expectedArgs);

      await lineaRollup.connect(securityCouncil).unsetVerifierAddress(2);

      const unsetVerifierCall = lineaRollup.connect(securityCouncil).unsetVerifierAddress(2);
      expectedArgs = [ADDRESS_ZERO, 2, securityCouncil.address, oldVerifierAddress];

      await expectEvent(lineaRollup, unsetVerifierCall, "VerifierAddressChanged", expectedArgs);
    });
  });

  describe("Data submission tests", () => {
    beforeEach(async () => {
      await lineaRollup.setLastFinalizedBlock(0);
    });

    const [DATA_ONE] = generateCallDataSubmission(0, 1);

    it("Fails when the compressed data is empty", async () => {
      const [submissionData] = generateCallDataSubmission(0, 1);
      submissionData.compressedData = EMPTY_CALLDATA;

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(submissionData, prevShnarf, secondExpectedShnarf, { gasLimit: 30_000_000 });
      await expectRevertWithCustomError(lineaRollup, submitDataCall, "EmptySubmissionData");
    });

    it("Should succesfully submit 1 compressed data chunk setting values", async () => {
      const [submissionData] = generateCallDataSubmission(0, 1);

      await expect(
        lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionData, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 }),
      ).to.not.be.reverted;

      const finalBlockNumber = await lineaRollup.shnarfFinalBlockNumbers(expectedShnarf);
      expect(finalBlockNumber).to.equal(submissionData.finalBlockInData);
    });

    it("Should successfully submit 2 compressed data chunks in two transactions", async () => {
      const [firstSubmissionData, secondSubmissionData] = generateCallDataSubmission(0, 2);

      await expect(
        lineaRollup
          .connect(operator)
          .submitDataAsCalldata(firstSubmissionData, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 }),
      ).to.not.be.reverted;

      await expect(
        lineaRollup.connect(operator).submitDataAsCalldata(secondSubmissionData, expectedShnarf, secondExpectedShnarf, {
          gasLimit: 30_000_000,
        }),
      ).to.not.be.reverted;

      const finalBlockNumber = await lineaRollup.shnarfFinalBlockNumbers(secondExpectedShnarf);

      expect(finalBlockNumber).to.equal(secondSubmissionData.finalBlockInData);
    });

    it("Should emit an event while submitting 1 compressed data chunk", async () => {
      const [submissionData] = generateCallDataSubmission(0, 1);

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(submissionData, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });
      const eventArgs = [expectedShnarf, 1, 46];

      await expectEvent(lineaRollup, submitDataCall, "DataSubmittedV2", eventArgs);
    });

    it("Should fail if the stored shnarf block number + 1 does not match the starting submission number", async () => {
      const [submissionData] = generateCallDataSubmission(0, 1);

      await lineaRollup.setShnarfFinalBlockNumber(prevShnarf, 99n);

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(submissionData, prevShnarf, secondExpectedShnarf, { gasLimit: 30_000_000 });
      const eventArgs = [100n, 1n];

      await expectRevertWithCustomError(lineaRollup, submitDataCall, "DataStartingBlockDoesNotMatch", eventArgs);
    });

    it("Should fail if the final state root hash is empty", async () => {
      const [submissionData] = generateCallDataSubmission(0, 1);

      submissionData.finalStateRootHash = HASH_ZERO;

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(submissionData, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });

      await expectRevertWithCustomError(lineaRollup, submitDataCall, "FinalBlockStateEqualsZeroHash");
    });

    it("Should fail if the block numbers are out of sequence", async () => {
      const [firstSubmissionData, secondSubmissionData] = generateCallDataSubmission(0, 2);

      await expect(
        lineaRollup
          .connect(operator)
          .submitDataAsCalldata(firstSubmissionData, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 }),
      ).to.not.be.reverted;

      const expectedFirstBlock = secondSubmissionData.firstBlockInData;
      secondSubmissionData.firstBlockInData = secondSubmissionData.firstBlockInData + 1n;

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(secondSubmissionData, expectedShnarf, secondExpectedShnarf, { gasLimit: 30_000_000 });
      const eventArgs = [expectedFirstBlock, secondSubmissionData.firstBlockInData];

      await expectRevertWithCustomError(lineaRollup, submitDataCall, "DataStartingBlockDoesNotMatch", eventArgs);
    });

    it("Should fail to submit where expected shnarf is wrong", async () => {
      const [firstSubmissionData, secondSubmissionData] = generateCallDataSubmission(0, 2);

      await expect(
        lineaRollup
          .connect(operator)
          .submitDataAsCalldata(firstSubmissionData, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 }),
      ).to.not.be.reverted;

      const wrongComputedShnarf = generateRandomBytes(32);

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(secondSubmissionData, expectedShnarf, wrongComputedShnarf, { gasLimit: 30_000_000 });

      const eventArgs = [wrongComputedShnarf, secondExpectedShnarf];

      await expectRevertWithCustomError(lineaRollup, submitDataCall, "FinalShnarfWrong", eventArgs);
    });

    it("Should revert if the caller does not have the OPERATOR_ROLE", async () => {
      const submitDataCall = lineaRollup
        .connect(nonAuthorizedAccount)
        .submitDataAsCalldata(DATA_ONE, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });

      await expectRevertWithReason(submitDataCall, buildAccessErrorMessage(nonAuthorizedAccount, OPERATOR_ROLE));
    });

    it("Should revert if GENERAL_PAUSE_TYPE is enabled", async () => {
      await lineaRollup.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(DATA_ONE, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });

      await expectRevertWithCustomError(lineaRollup, submitDataCall, "IsPaused", [GENERAL_PAUSE_TYPE]);
    });

    it("Should revert if CALLDATA_SUBMISSION_PAUSE_TYPE is enabled", async () => {
      await lineaRollup.connect(securityCouncil).pauseByType(CALLDATA_SUBMISSION_PAUSE_TYPE);

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(DATA_ONE, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });

      await expectRevertWithCustomError(lineaRollup, submitDataCall, "IsPaused", [CALLDATA_SUBMISSION_PAUSE_TYPE]);
    });

    it("Should revert with FirstBlockLessThanOrEqualToLastFinalizedBlock when submitting data with firstBlockInData less than currentL2BlockNumber", async () => {
      await lineaRollup.setLastFinalizedBlock(1_000_000);

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(DATA_ONE, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });

      await expectRevertWithCustomError(lineaRollup, submitDataCall, "FirstBlockLessThanOrEqualToLastFinalizedBlock", [
        DATA_ONE.firstBlockInData,
        1000000,
      ]);
    });

    it("Should revert with FirstBlockGreaterThanFinalBlock when submitting data with firstBlockInData greater than finalBlockInData", async () => {
      const submissionData: CalldataSubmissionData = {
        ...DATA_ONE,
        firstBlockInData: DATA_ONE.firstBlockInData,
        finalBlockInData: DATA_ONE.firstBlockInData - 1n,
      };

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(submissionData, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });

      await expectRevertWithCustomError(lineaRollup, submitDataCall, "FirstBlockGreaterThanFinalBlock", [
        submissionData.firstBlockInData,
        submissionData.finalBlockInData,
      ]);
    });

    it("Should revert with DataAlreadySubmitted when submitting same compressed data twice in 2 separate transactions", async () => {
      await lineaRollup
        .connect(operator)
        .submitDataAsCalldata(DATA_ONE, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(DATA_ONE, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });

      await expectRevertWithCustomError(lineaRollup, submitDataCall, "DataAlreadySubmitted", [expectedShnarf]);
    });

    it("Should revert with DataAlreadySubmitted when submitting same data, differing block numbers", async () => {
      await lineaRollup
        .connect(operator)
        .submitDataAsCalldata(DATA_ONE, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });

      const [dataOneCopy] = generateCallDataSubmission(0, 1);
      dataOneCopy.finalBlockInData = 234253242n;

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(dataOneCopy, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });

      await expectRevertWithCustomError(lineaRollup, submitDataCall, "DataAlreadySubmitted", [expectedShnarf]);
    });

    it("Should revert with SnarkHashIsZeroHash when snarkHash is zero hash", async () => {
      const submissionData: CalldataSubmissionData = {
        ...DATA_ONE,
        snarkHash: HASH_ZERO,
      };

      const submitDataCall = lineaRollup
        .connect(operator)
        .submitDataAsCalldata(submissionData, prevShnarf, expectedShnarf, { gasLimit: 30_000_000 });

      await expectRevertWithCustomError(lineaRollup, submitDataCall, "SnarkHashIsZeroHash");
    });
  });

  describe("EIP-4844 Blob submission tests", () => {
    beforeEach(async () => {
      await lineaRollup.setLastFinalizedBlock(0);
      await lineaRollup.setupParentShnarf(prevShnarf, 0);
      await lineaRollup.setupParentDataShnarf(parentDataHash, prevShnarf);
      await lineaRollup.setupParentFinalizedStateRoot(parentDataHash, parentStateRootHash);
    });

    it("Should successfully submit blobs", async () => {
      const operatorHDSigner = getWalletForIndex(2);
      const lineaRollupAddress = await lineaRollup.getAddress();
      const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

      const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
        blobDataSubmission,
        parentShnarf,
        finalShnarf,
      ]);

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
      const nonce = await operatorHDSigner.getNonce();

      const transaction = Transaction.from({
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: (await ethers.provider.getNetwork()).chainId,
        type: 3,
        nonce,
        value: 0,
        gasLimit: 5_000_000,
        kzg,
        maxFeePerBlobGas: 1n,
        blobs: compressedBlobs,
      });

      const signedTx = await operatorHDSigner.signTransaction(transaction);

      await ethers.provider.broadcastTransaction(signedTx);

      const finalBlockNumber = await lineaRollup.shnarfFinalBlockNumbers(finalShnarf);
      expect(finalBlockNumber).to.equal(blobDataSubmission[0].submissionData.finalBlockInData);
    });

    it("Fails when the blob submission data is missing", async () => {
      const operatorHDSigner = getWalletForIndex(2);

      const lineaRollupAddress = await lineaRollup.getAddress();
      const { compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

      const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [[], parentShnarf, finalShnarf]);

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
      const nonce = await operatorHDSigner.getNonce();

      const transaction = Transaction.from({
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: (await ethers.provider.getNetwork()).chainId,
        type: 3,
        nonce,
        value: 0,
        gasLimit: 5_000_000,
        kzg,
        maxFeePerBlobGas: 1n,
        blobs: compressedBlobs,
      });

      const signedTx = await operatorHDSigner.signTransaction(transaction);

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

    it("Should revert if GENERAL_PAUSE_TYPE is enabled", async () => {
      const { blobDataSubmission, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

      await lineaRollup.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      await expectRevertWithCustomError(
        lineaRollup,
        lineaRollup.connect(operator).submitBlobs(blobDataSubmission, parentShnarf, finalShnarf),
        "IsPaused",
        [GENERAL_PAUSE_TYPE],
      );
    });

    it("Should revert if BLOB_SUBMISSION_PAUSE_TYPE is enabled", async () => {
      const { blobDataSubmission, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

      await lineaRollup.connect(securityCouncil).pauseByType(BLOB_SUBMISSION_PAUSE_TYPE);

      await expectRevertWithCustomError(
        lineaRollup,
        lineaRollup.connect(operator).submitBlobs(blobDataSubmission, parentShnarf, finalShnarf),
        "IsPaused",
        [BLOB_SUBMISSION_PAUSE_TYPE],
      );
    });

    it("Should revert if the blob data is empty at any index", async () => {
      const operatorHDSigner = getWalletForIndex(2);
      const lineaRollupAddress = await lineaRollup.getAddress();
      const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 2);

      const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
        blobDataSubmission,
        parentShnarf,
        finalShnarf,
      ]);

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
      const nonce = await operatorHDSigner.getNonce();

      const transaction = Transaction.from({
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: (await ethers.provider.getNetwork()).chainId,
        type: 3,
        nonce,
        value: 0,
        gasLimit: 5_000_000,
        kzg,
        maxFeePerBlobGas: 1n,
        blobs: [compressedBlobs[0]],
      });

      const signedTx = await operatorHDSigner.signTransaction(transaction);

      await expectRevertWithCustomError(
        lineaRollup,
        ethers.provider.broadcastTransaction(signedTx),
        "EmptyBlobDataAtIndex",
        [1n],
      );
    });
    it("Should fail if the final state root hash is empty", async () => {
      const operatorHDSigner = getWalletForIndex(2);

      const lineaRollupAddress = await lineaRollup.getAddress();
      const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

      blobDataSubmission[0].submissionData.finalStateRootHash = HASH_ZERO;

      const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
        blobDataSubmission,
        parentShnarf,
        finalShnarf,
      ]);

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
      const nonce = await operatorHDSigner.getNonce();

      const transaction = Transaction.from({
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: (await ethers.provider.getNetwork()).chainId,
        type: 3,
        nonce,
        value: 0,
        gasLimit: 5_000_000,
        kzg,
        maxFeePerBlobGas: 1n,
        blobs: compressedBlobs,
      });

      const signedTx = await operatorHDSigner.signTransaction(transaction);

      await expectRevertWithCustomError(
        lineaRollup,
        ethers.provider.broadcastTransaction(signedTx),
        "FinalBlockStateEqualsZeroHash",
      );
    });

    it("Should revert with SnarkHashIsZeroHash when snarkHash is zero hash", async () => {
      const operatorHDSigner = getWalletForIndex(2);

      const lineaRollupAddress = await lineaRollup.getAddress();
      const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

      // Set the snarkHash to HASH_ZERO for a specific index
      const emptyDataIndex = 0;
      blobDataSubmission[emptyDataIndex].submissionData.snarkHash = HASH_ZERO;

      const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
        blobDataSubmission,
        parentShnarf,
        finalShnarf,
      ]);

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
      const nonce = await operatorHDSigner.getNonce();

      const transaction = Transaction.from({
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: (await ethers.provider.getNetwork()).chainId,
        type: 3,
        nonce,
        value: 0,
        gasLimit: 5_000_000,
        kzg,
        maxFeePerBlobGas: 1n,
        blobs: compressedBlobs,
      });

      const signedTx = await operatorHDSigner.signTransaction(transaction);

      await expectRevertWithCustomError(
        lineaRollup,
        ethers.provider.broadcastTransaction(signedTx),
        "SnarkHashIsZeroHash",
      );
    });

    it("Should fail if the block numbers are out of sequence", async () => {
      const operatorHDSigner = getWalletForIndex(2);

      const lineaRollupAddress = await lineaRollup.getAddress();
      const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 2);

      blobDataSubmission[1].submissionData.firstBlockInData =
        blobDataSubmission[1].submissionData.finalBlockInData + 1n;

      const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
        blobDataSubmission,
        parentShnarf,
        finalShnarf,
      ]);

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
      const nonce = await operatorHDSigner.getNonce();

      const transaction = Transaction.from({
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: (await ethers.provider.getNetwork()).chainId,
        type: 3,
        nonce,
        value: 0,
        gasLimit: 5_000_000,
        kzg,
        maxFeePerBlobGas: 1n,
        blobs: compressedBlobs,
      });

      const signedTx = await operatorHDSigner.signTransaction(transaction);

      await expectRevertWithCustomError(
        lineaRollup,
        ethers.provider.broadcastTransaction(signedTx),
        "DataStartingBlockDoesNotMatch",
        [
          blobDataSubmission[0].submissionData.finalBlockInData + 1n,
          blobDataSubmission[1].submissionData.firstBlockInData,
        ],
      );
    });

    it("Should revert with FirstBlockLessThanOrEqualToLastFinalizedBlock when submitting data with firstBlockInData less than currentL2BlockNumber", async () => {
      const operatorHDSigner = getWalletForIndex(2);

      const lineaRollupAddress = await lineaRollup.getAddress();
      const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

      // Set the currentL2BlockNumber to a value greater than the firstBlockInData
      const currentL2BlockNumber = 1_000_000n;
      await lineaRollup.setLastFinalizedBlock(currentL2BlockNumber);

      const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
        blobDataSubmission,
        parentShnarf,
        finalShnarf,
      ]);

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
      const nonce = await operatorHDSigner.getNonce();

      const transaction = Transaction.from({
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: (await ethers.provider.getNetwork()).chainId,
        type: 3,
        nonce,
        value: 0,
        gasLimit: 5_000_000,
        kzg,
        maxFeePerBlobGas: 1n,
        blobs: compressedBlobs,
      });

      const signedTx = await operatorHDSigner.signTransaction(transaction);

      await expectRevertWithCustomError(
        lineaRollup,
        ethers.provider.broadcastTransaction(signedTx),
        "FirstBlockLessThanOrEqualToLastFinalizedBlock",
        [blobDataSubmission[0].submissionData.firstBlockInData, currentL2BlockNumber],
      );
    });

    it("Should revert with FirstBlockGreaterThanFinalBlock when submitting data with firstBlockInData greater than finalBlockInData", async () => {
      const operatorHDSigner = getWalletForIndex(2);

      const lineaRollupAddress = await lineaRollup.getAddress();
      const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

      // Set the firstBlockInData to be greater than the finalBlockInData
      blobDataSubmission[0].submissionData.finalBlockInData =
        blobDataSubmission[0].submissionData.firstBlockInData - 1n;

      const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
        blobDataSubmission,
        parentShnarf,
        finalShnarf,
      ]);

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
      const nonce = await operatorHDSigner.getNonce();

      const transaction = Transaction.from({
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: (await ethers.provider.getNetwork()).chainId,
        type: 3,
        nonce,
        value: 0,
        gasLimit: 5_000_000,
        kzg,
        maxFeePerBlobGas: 1n,
        blobs: compressedBlobs,
      });

      const signedTx = await operatorHDSigner.signTransaction(transaction);

      await expectRevertWithCustomError(
        lineaRollup,
        ethers.provider.broadcastTransaction(signedTx),
        "FirstBlockGreaterThanFinalBlock",
        [blobDataSubmission[0].submissionData.firstBlockInData, blobDataSubmission[0].submissionData.finalBlockInData],
      );
    });

    it("Should revert if the final shnarf is wrong", async () => {
      const operatorHDSigner = getWalletForIndex(2);
      const lineaRollupAddress = await lineaRollup.getAddress();
      const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 2);
      const badFinalShnarf = generateRandomBytes(32);

      const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
        blobDataSubmission,
        parentShnarf,
        badFinalShnarf,
      ]);

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
      const nonce = await operatorHDSigner.getNonce();

      const transaction = Transaction.from({
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: (await ethers.provider.getNetwork()).chainId,
        type: 3,
        nonce,
        value: 0,
        gasLimit: 5_000_000,
        kzg,
        maxFeePerBlobGas: 1n,
        blobs: compressedBlobs,
      });

      const signedTx = await operatorHDSigner.signTransaction(transaction);

      await expectRevertWithCustomError(
        lineaRollup,
        ethers.provider.broadcastTransaction(signedTx),
        "FinalShnarfWrong",
        [badFinalShnarf, finalShnarf],
      );
    });

    it("Should revert if the data has already been submitted", async () => {
      await sendBlobTransaction(0, 1);

      const operatorHDSigner = getWalletForIndex(2);

      const lineaRollupAddress = await lineaRollup.getAddress();
      const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

      // Try to submit the same blob data again
      const encodedCall2 = lineaRollup.interface.encodeFunctionData("submitBlobs", [
        blobDataSubmission,
        parentShnarf,
        finalShnarf,
      ]);

      const { maxFeePerGas: maxFeePerGas2, maxPriorityFeePerGas: maxPriorityFeePerGas2 } =
        await ethers.provider.getFeeData();
      const nonce2 = await operatorHDSigner.getNonce();

      const transaction2 = Transaction.from({
        data: encodedCall2,
        maxPriorityFeePerGas: maxPriorityFeePerGas2!,
        maxFeePerGas: maxFeePerGas2!,
        to: lineaRollupAddress,
        chainId: (await ethers.provider.getNetwork()).chainId,
        type: 3,
        nonce: nonce2,
        value: 0,
        gasLimit: 5_000_000,
        kzg,
        maxFeePerBlobGas: 1n,
        blobs: compressedBlobs,
      });

      const signedTx2 = await operatorHDSigner.signTransaction(transaction2);

      await expectRevertWithCustomError(
        lineaRollup,
        ethers.provider.broadcastTransaction(signedTx2),
        "DataAlreadySubmitted",
        [finalShnarf],
      );
    });

    it("Should revert with PointEvaluationFailed when point evaluation fails", async () => {
      const operatorHDSigner = getWalletForIndex(2);

      const lineaRollupAddress = await lineaRollup.getAddress();
      const { blobDataSubmission, compressedBlobs, parentShnarf, finalShnarf } = generateBlobDataSubmission(0, 1);

      // Modify the kzgProof to an invalid value to trigger the PointEvaluationFailed revert
      blobDataSubmission[0].kzgProof = HASH_ZERO;

      const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
        blobDataSubmission,
        parentShnarf,
        finalShnarf,
      ]);

      const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
      const nonce = await operatorHDSigner.getNonce();

      const transaction = Transaction.from({
        data: encodedCall,
        maxPriorityFeePerGas: maxPriorityFeePerGas!,
        maxFeePerGas: maxFeePerGas!,
        to: lineaRollupAddress,
        chainId: (await ethers.provider.getNetwork()).chainId,
        type: 3,
        nonce,
        value: 0,
        gasLimit: 5_000_000,
        kzg,
        maxFeePerBlobGas: 1n,
        blobs: compressedBlobs,
      });

      const signedTx = await operatorHDSigner.signTransaction(transaction);

      await expectRevertWithCustomError(
        lineaRollup,
        ethers.provider.broadcastTransaction(signedTx),
        "PointEvaluationFailed",
      );
    });

    it("Should submit 2 blobs, then submit another 2 blobs and finalize", async () => {
      // Submit 2 blobs
      await sendBlobTransaction(0, 2);
      // Submit another 2 blobs
      await sendBlobTransaction(2, 4);
      // Finalize 4 blobs
      await expectSuccessfulFinalizeWithProof(
        blobAggregatedProof1To155,
        4,
        fourthCompressedDataContent.finalStateRootHash,
        generateBlobParentShnarfData,
      );
    });

    it("Should fail to finalize with not enough gas for the rollup (pre-verifier)", async () => {
      // Submit 2 blobs
      await sendBlobTransaction(0, 2);
      // Submit another 2 blobs
      await sendBlobTransaction(2, 4);

      // Finalize 4 blobs
      const finalizationData = await generateFinalizationData({
        l1RollingHash: blobAggregatedProof1To155.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(blobAggregatedProof1To155.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(blobAggregatedProof1To155.parentAggregationLastBlockTimestamp),
        finalBlockInData: BigInt(blobAggregatedProof1To155.finalBlockNumber),
        parentStateRootHash: blobAggregatedProof1To155.parentStateRootHash,
        finalTimestamp: BigInt(blobAggregatedProof1To155.finalTimestamp),
        l2MerkleRoots: blobAggregatedProof1To155.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(blobAggregatedProof1To155.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: blobAggregatedProof1To155.l2MessagingBlocksOffsets,
        aggregatedProof: blobAggregatedProof1To155.aggregatedProof,
        shnarfData: generateBlobParentShnarfData(4, false),
      });
      finalizationData.lastFinalizedL1RollingHash = HASH_ZERO;
      finalizationData.lastFinalizedL1RollingHashMessageNumber = 0n;

      finalizationData.lastFinalizedShnarf = blobAggregatedProof1To155.parentAggregationFinalShnarf;

      await lineaRollup.setRollingHash(
        blobAggregatedProof1To155.l1RollingHashMessageNumber,
        blobAggregatedProof1To155.l1RollingHash,
      );

      const finalizeCompressedCall = lineaRollup
        .connect(operator)
        .finalizeBlocksWithProof(
          blobAggregatedProof1To155.aggregatedProof,
          TEST_PUBLIC_VERIFIER_INDEX,
          finalizationData,
          { gasLimit: 50000 },
        );

      // there is no reason
      await expect(finalizeCompressedCall).to.be.reverted;
    });

    it("Should fail to finalize with not enough gas to verify", async () => {
      // Submit 2 blobs
      await sendBlobTransaction(0, 2);
      // Submit another 2 blobs
      await sendBlobTransaction(2, 4);

      // Finalize 4 blobs
      const finalizationData = await generateFinalizationData({
        l1RollingHash: blobAggregatedProof1To155.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(blobAggregatedProof1To155.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(blobAggregatedProof1To155.parentAggregationLastBlockTimestamp),
        finalBlockInData: BigInt(blobAggregatedProof1To155.finalBlockNumber),
        parentStateRootHash: blobAggregatedProof1To155.parentStateRootHash,
        finalTimestamp: BigInt(blobAggregatedProof1To155.finalTimestamp),
        l2MerkleRoots: blobAggregatedProof1To155.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(blobAggregatedProof1To155.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: blobAggregatedProof1To155.l2MessagingBlocksOffsets,
        aggregatedProof: blobAggregatedProof1To155.aggregatedProof,
        shnarfData: generateBlobParentShnarfData(4, false),
      });
      finalizationData.lastFinalizedL1RollingHash = HASH_ZERO;
      finalizationData.lastFinalizedL1RollingHashMessageNumber = 0n;

      finalizationData.lastFinalizedShnarf = blobAggregatedProof1To155.parentAggregationFinalShnarf;

      await lineaRollup.setRollingHash(
        blobAggregatedProof1To155.l1RollingHashMessageNumber,
        blobAggregatedProof1To155.l1RollingHash,
      );

      const finalizeCompressedCall = lineaRollup
        .connect(operator)
        .finalizeBlocksWithProof(
          blobAggregatedProof1To155.aggregatedProof,
          TEST_PUBLIC_VERIFIER_INDEX,
          finalizationData,
          { gasLimit: 400000 },
        );

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
        await deployRevertingVerifier(revertScenario);
        await lineaRollup.connect(securityCouncil).setVerifierAddress(revertingVerifier, 0);

        // Submit 2 blobs
        await sendBlobTransaction(0, 2);
        // Submit another 2 blobs
        await sendBlobTransaction(2, 4);

        // Finalize 4 blobs
        const finalizationData = await generateFinalizationData({
          l1RollingHash: blobAggregatedProof1To155.l1RollingHash,
          l1RollingHashMessageNumber: BigInt(blobAggregatedProof1To155.l1RollingHashMessageNumber),
          lastFinalizedTimestamp: BigInt(blobAggregatedProof1To155.parentAggregationLastBlockTimestamp),
          finalBlockInData: BigInt(blobAggregatedProof1To155.finalBlockNumber),
          parentStateRootHash: blobAggregatedProof1To155.parentStateRootHash,
          finalTimestamp: BigInt(blobAggregatedProof1To155.finalTimestamp),
          l2MerkleRoots: blobAggregatedProof1To155.l2MerkleRoots,
          l2MerkleTreesDepth: BigInt(blobAggregatedProof1To155.l2MerkleTreesDepth),
          l2MessagingBlocksOffsets: blobAggregatedProof1To155.l2MessagingBlocksOffsets,
          aggregatedProof: blobAggregatedProof1To155.aggregatedProof,
          shnarfData: generateBlobParentShnarfData(4, false),
        });
        finalizationData.lastFinalizedL1RollingHash = HASH_ZERO;
        finalizationData.lastFinalizedL1RollingHashMessageNumber = 0n;

        finalizationData.lastFinalizedShnarf = blobAggregatedProof1To155.parentAggregationFinalShnarf;

        await lineaRollup.setRollingHash(
          blobAggregatedProof1To155.l1RollingHashMessageNumber,
          blobAggregatedProof1To155.l1RollingHash,
        );

        const finalizeCompressedCall = lineaRollup
          .connect(operator)
          .finalizeBlocksWithProof(
            blobAggregatedProof1To155.aggregatedProof,
            TEST_PUBLIC_VERIFIER_INDEX,
            finalizationData,
            { gasLimit: 400000 },
          );

        await expectRevertWithCustomError(
          lineaRollup,
          finalizeCompressedCall,
          "InvalidProofOrProofVerificationRanOutOfGas",
          ["Unknown"],
        );
      });
    });

    it("Should successfully submit 2 blobs twice then finalize in two separate finalizations", async () => {
      // Submit 2 blobs
      await sendBlobTransaction(0, 2, true);
      // Submit another 2 blobs
      await sendBlobTransaction(2, 4, true);
      // Finalize first 2 blobs
      await expectSuccessfulFinalizeWithProof(
        blobMultipleAggregatedProof1To81,
        2,
        secondCompressedDataContent.finalStateRootHash,
        generateBlobParentShnarfData,
        true,
      );
      // Finalize last 2 blobs
      await expectSuccessfulFinalizeWithProof(
        blobMultipleAggregatedProof82To153,
        4,
        fourthMultipleBlobDataContent.finalStateRootHash,
        generateBlobParentShnarfData,
        true,
        blobMultipleAggregatedProof1To81.l1RollingHash,
        BigInt(blobMultipleAggregatedProof1To81.l1RollingHashMessageNumber),
      );
    });
  });

  describe("Blocks finalization without proof", () => {
    const messageHash = generateRandomBytes(32);

    beforeEach(async () => {
      await lineaRollup.addRollingHash(10, messageHash);
      await lineaRollup.setLastFinalizedBlock(0);
    });

    describe("With and without submission data", () => {
      it("Should revert if caller does not the role 'FINALIZE_WITHOUT_PROOF_ROLE'", async () => {
        const finalizationData = await generateFinalizationData();

        const finalizeCall = lineaRollup.connect(operator).finalizeBlocksWithoutProof(finalizationData);

        await expectRevertWithReason(finalizeCall, buildAccessErrorMessage(operator, FINALIZE_WITHOUT_PROOF_ROLE));
      });

      it("Should revert if GENERAL_PAUSE_TYPE is enabled", async () => {
        const finalizationData = await generateFinalizationData();

        await lineaRollup.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

        const finalizeCall = lineaRollup.connect(securityCouncil).finalizeBlocksWithoutProof(finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCall, "IsPaused", [GENERAL_PAUSE_TYPE]);
      });

      it("Should revert if _finalizationData.finalBlockNumber is less than or equal to currentL2BlockNumber", async () => {
        await lineaRollup.setLastFinalizedBlock(10_000_000);

        const finalizationData = await generateFinalizationData();

        // finalization block is set to 0 and the hash is zero - the test is to perform other validations
        finalizationData.parentStateRootHash = HASH_ZERO;

        const finalizeCall = lineaRollup.connect(securityCouncil).finalizeBlocksWithoutProof(finalizationData);

        await expectRevertWithCustomError(
          lineaRollup,
          finalizeCall,
          "FinalBlockNumberLessThanOrEqualToLastFinalizedBlock",
          [finalizationData.finalBlockInData, 10_000_000],
        );
      });

      it("Should revert if l1 message number == 0 and l1 rolling hash is not empty", async () => {
        const finalizationData = await generateFinalizationData({
          l1RollingHashMessageNumber: 0n,
        });

        // finalization block is set to 0 and the hash is zero - the test is to perform other validations
        finalizationData.parentStateRootHash = firstCompressedDataContent.parentStateRootHash;

        const finalizeCall = lineaRollup.connect(securityCouncil).finalizeBlocksWithoutProof(finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCall, "MissingMessageNumberForRollingHash", [
          finalizationData.l1RollingHash,
        ]);
      });

      it("Should revert if l1 message number != 0 and l1 rolling hash is empty", async () => {
        const finalizationData = await generateFinalizationData({ l1RollingHash: HASH_ZERO });

        // finalization block is set to 0 and the hash is zero - the test is to perform other validations
        finalizationData.parentStateRootHash = HASH_ZERO;

        const finalizeCall = lineaRollup.connect(securityCouncil).finalizeBlocksWithoutProof(finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCall, "MissingRollingHashForMessageNumber", [
          finalizationData.l1RollingHashMessageNumber,
        ]);
      });

      it("Should revert if l1RollingHash does not exist on L1", async () => {
        const finalizationData = await generateFinalizationData();

        // finalization block is set to 0 and the hash is zero - the test is to perform other validations
        finalizationData.parentStateRootHash = HASH_ZERO;

        const finalizeCall = lineaRollup.connect(securityCouncil).finalizeBlocksWithoutProof(finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCall, "L1RollingHashDoesNotExistOnL1", [
          finalizationData.l1RollingHashMessageNumber,
          finalizationData.l1RollingHash,
        ]);
      });

      it("Should revert if timestamps are not in sequence", async () => {
        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
        });

        // finalization block is set to 0 and the hash is zero - the test is to perform other validations
        finalizationData.parentStateRootHash = HASH_ZERO;
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
            1683325137n,
          ],
        );

        const finalizeCompressedCall = lineaRollup
          .connect(securityCouncil)
          .finalizeBlocksWithoutProof(finalizationData);
        await expectRevertWithCustomError(lineaRollup, finalizeCompressedCall, "FinalizationStateIncorrect", [
          expectedHashValue,
          actualHashValue,
        ]);
      });

      it("Should revert if finalizationData.finalTimestamp is greater than the block.timestamp", async () => {
        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
          lastFinalizedTimestamp: 1683325137n,
          finalTimestamp: BigInt(new Date(new Date().setHours(new Date().getHours() + 2)).getTime()),
        });

        // finalization block is set to 0 and the hash is zero - the test is to perform other validations
        finalizationData.parentStateRootHash = HASH_ZERO;

        const finalizeCompressedCall = lineaRollup
          .connect(securityCouncil)
          .finalizeBlocksWithoutProof(finalizationData);
        await expectRevertWithCustomError(lineaRollup, finalizeCompressedCall, "FinalizationInTheFuture", [
          finalizationData.finalTimestamp,
          (await networkTime.latest()) + 1,
        ]);
      });

      it("Should revert if the parent datahash's fingerprint does not match", async () => {
        const [submissionDataBeforeFinalization] = generateCallDataSubmission(0, 1);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionDataBeforeFinalization, prevShnarf, expectedShnarf, {
            gasLimit: 30_000_000,
          });

        const finalSubmissionData = generateParentSubmissionDataForIndex(1);

        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
          lastFinalizedTimestamp: 1683325137n,
          finalBlockInData: BigInt(100n),
          shnarfData: generateParentShnarfData(0),
        });

        finalSubmissionData.shnarf = generateRandomBytes(32);

        const finalizeCompressedCall = lineaRollup
          .connect(securityCouncil)
          .finalizeBlocksWithoutProof(finalizationData);
        await expectRevertWithCustomError(
          lineaRollup,
          finalizeCompressedCall,
          "FinalBlockDoesNotMatchShnarfFinalBlock",
          [finalizationData.finalBlockInData, await lineaRollup.dataShnarfHashes(finalSubmissionData.shnarf)],
        );
      });
    });

    describe("Without submission data", () => {
      it("Should revert with if the final block state equals the zero hash", async () => {
        const submissionDataBeforeFinalization = generateCallDataSubmission(0, 2);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionDataBeforeFinalization[0], prevShnarf, expectedShnarf, {
            gasLimit: 30_000_000,
          });

        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionDataBeforeFinalization[1], expectedShnarf, secondExpectedShnarf, {
            gasLimit: 30_000_000,
          });

        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
          lastFinalizedTimestamp: 1683325137n,
        });

        finalizationData.shnarfData.finalStateRootHash = HASH_ZERO;

        const finalizeCall = lineaRollup.connect(securityCouncil).finalizeBlocksWithoutProof(finalizationData);

        await expectRevertWithCustomError(lineaRollup, finalizeCall, "FinalBlockStateEqualsZeroHash");
      });

      it("Should successfully finalize blocks and emit DataFinalized event", async () => {
        const submissionDataBeforeFinalization = generateCallDataSubmission(0, 2);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionDataBeforeFinalization[0], prevShnarf, expectedShnarf, {
            gasLimit: 30_000_000,
          });

        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionDataBeforeFinalization[1], expectedShnarf, secondExpectedShnarf, {
            gasLimit: 30_000_000,
          });

        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
          lastFinalizedTimestamp: 1683325137n,
          finalBlockInData: BigInt(submissionDataBeforeFinalization[1].finalBlockInData),
          parentStateRootHash: parentStateRootHash,
          shnarfData: generateParentShnarfData(2),
        });

        const finalizeCompressedCall = lineaRollup
          .connect(securityCouncil)
          .finalizeBlocksWithoutProof(finalizationData);
        const eventArgs = [
          finalizationData.finalBlockInData,
          finalizationData.parentStateRootHash,
          finalizationData.shnarfData.finalStateRootHash,
          false,
        ];

        await expectEvent(lineaRollup, finalizeCompressedCall, "DataFinalized", eventArgs);
      });

      it("Should successfully finalize blocks and store the last state root hash, the final timestamp, the final block number", async () => {
        const submissionDataBeforeFinalization = generateCallDataSubmission(0, 2);

        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionDataBeforeFinalization[0], prevShnarf, expectedShnarf, {
            gasLimit: 30_000_000,
          });

        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionDataBeforeFinalization[1], expectedShnarf, secondExpectedShnarf, {
            gasLimit: 30_000_000,
          });

        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
          lastFinalizedTimestamp: 1683325137n,
          finalBlockInData: submissionDataBeforeFinalization[1].finalBlockInData,
          parentStateRootHash: parentStateRootHash,
          shnarfData: generateParentShnarfData(2),
        });

        expect(await lineaRollup.connect(securityCouncil).finalizeBlocksWithoutProof(finalizationData)).to.not.be
          .reverted;

        const [finalStateRootHash, lastFinalizedBlockNumber, lastFinalizedState] = await Promise.all([
          lineaRollup.stateRootHashes(finalizationData.finalBlockInData),
          lineaRollup.currentL2BlockNumber(),
          lineaRollup.currentFinalizedState(),
        ]);

        expect(finalStateRootHash).to.equal(finalizationData.shnarfData.finalStateRootHash);
        expect(lastFinalizedBlockNumber).to.equal(finalizationData.finalBlockInData);
        expect(lastFinalizedState).to.equal(
          generateKeccak256(
            ["uint256", "bytes32", "uint256"],
            [
              finalizationData.l1RollingHashMessageNumber,
              finalizationData.l1RollingHash,
              finalizationData.finalTimestamp,
            ],
          ),
        );
      });

      it("Should successfully finalize blocks and anchor L2 merkle root, emit an event for each L2 block containing L2->L1 messages", async () => {
        const submissionDataBeforeFinalization = generateCallDataSubmission(0, 2);

        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionDataBeforeFinalization[0], prevShnarf, expectedShnarf, {
            gasLimit: 30_000_000,
          });

        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionDataBeforeFinalization[1], expectedShnarf, secondExpectedShnarf, {
            gasLimit: 30_000_000,
          });

        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
          lastFinalizedTimestamp: 1683325137n,
          finalBlockInData: submissionDataBeforeFinalization[1].finalBlockInData,
          parentStateRootHash: parentStateRootHash,
          shnarfData: generateParentShnarfData(2),
        });

        const currentL2BlockNumber = await lineaRollup.currentL2BlockNumber();

        const tx = await lineaRollup.connect(securityCouncil).finalizeBlocksWithoutProof(finalizationData);
        await tx.wait();

        const events = await lineaRollup.queryFilter(lineaRollup.filters.L2MessagingBlockAnchored());

        expect(events.length).to.equal(1);

        for (let i = 0; i < events.length; i++) {
          expect(events[i].args?.l2Block).to.deep.equal(
            currentL2BlockNumber + BigInt(`0x${finalizationData.l2MessagingBlocksOffsets.slice(i * 4 + 2, i * 4 + 6)}`),
          );
        }

        for (let i = 0; i < finalizationData.l2MerkleRoots.length; i++) {
          const l2MerkleRootTreeDepth = await lineaRollup.l2MerkleRootsDepths(finalizationData.l2MerkleRoots[i]);
          expect(l2MerkleRootTreeDepth).to.equal(finalizationData.l2MerkleTreesDepth);
        }
      });

      it("Should successfully finalize blocks when we submit data1 and data2 but only finalizing data1", async () => {
        const submissionDataBeforeFinalization = generateCallDataSubmission(0, 2);

        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionDataBeforeFinalization[0], prevShnarf, expectedShnarf, {
            gasLimit: 30_000_000,
          });

        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(submissionDataBeforeFinalization[1], expectedShnarf, secondExpectedShnarf, {
            gasLimit: 30_000_000,
          });

        const finalizationData = await generateFinalizationData({
          l1RollingHash: calculateRollingHash(HASH_ZERO, messageHash),
          l1RollingHashMessageNumber: 10n,
          lastFinalizedTimestamp: 1683325137n,
          finalBlockInData: submissionDataBeforeFinalization[0].finalBlockInData,
          parentStateRootHash: parentStateRootHash,
          shnarfData: generateParentShnarfData(1),
        });

        expect(await lineaRollup.connect(securityCouncil).finalizeBlocksWithoutProof(finalizationData)).to.not.be
          .reverted;

        const [finalStateRootHash, lastFinalizedBlockNumber, lastFinalizedState] = await Promise.all([
          lineaRollup.stateRootHashes(finalizationData.finalBlockInData),
          lineaRollup.currentL2BlockNumber(),
          lineaRollup.currentFinalizedState(),
        ]);

        expect(finalStateRootHash).to.equal(finalizationData.shnarfData.finalStateRootHash);
        expect(lastFinalizedBlockNumber).to.equal(finalizationData.finalBlockInData);
        expect(lastFinalizedState).to.equal(
          generateKeccak256(
            ["uint256", "bytes32", "uint256"],
            [
              finalizationData.l1RollingHashMessageNumber,
              finalizationData.l1RollingHash,
              finalizationData.finalTimestamp,
            ],
          ),
        );
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
        .finalizeBlocksWithProof(
          calldataAggregatedProof1To155.aggregatedProof,
          TEST_PUBLIC_VERIFIER_INDEX,
          finalizationData,
        );
      await expectRevertWithReason(finalizeCall, buildAccessErrorMessage(nonAuthorizedAccount, OPERATOR_ROLE));
    });

    it("Should revert if GENERAL_PAUSE_TYPE is enabled", async () => {
      const finalizationData = await generateFinalizationData();

      await lineaRollup.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocksWithProof(EMPTY_CALLDATA, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "IsPaused", [GENERAL_PAUSE_TYPE]);
    });

    it("Should revert if FINALIZATION_PAUSE_TYPE is enabled", async () => {
      const finalizationData = await generateFinalizationData();

      await lineaRollup.connect(securityCouncil).pauseByType(FINALIZATION_PAUSE_TYPE);

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocksWithProof(EMPTY_CALLDATA, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "IsPaused", [FINALIZATION_PAUSE_TYPE]);
    });

    it("Should revert if the proof is empty", async () => {
      const finalizationData = await generateFinalizationData();

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocksWithProof(EMPTY_CALLDATA, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
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
            gasLimit: 30_000_000,
          });
        index++;
      }

      const finalizationData = await generateFinalizationData({
        lastFinalizedTimestamp: 1683325137n,
        parentStateRootHash: generateRandomBytes(32),
        aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
      });

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocksWithProof(
          calldataAggregatedProof1To155.aggregatedProof,
          TEST_PUBLIC_VERIFIER_INDEX,
          finalizationData,
          {
            gasLimit: 30_000_000,
          },
        );
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
            gasLimit: 30_000_000,
          });
        index++;
      }

      await expectSuccessfulFinalizeWithProof(
        calldataAggregatedProof1To155,
        index,
        fourthCompressedDataContent.finalStateRootHash,
        generateParentShnarfData,
      );
    });

    it("Should revert if last finalized shnarf is wrong", async () => {
      // Submit 4 sets of compressed data setting the correct shnarf in storage
      const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);

      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: 30_000_000,
          });
        index++;
      }

      const finalizationData = await generateFinalizationData({
        l1RollingHash: calldataAggregatedProof1To155.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(calldataAggregatedProof1To155.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(calldataAggregatedProof1To155.parentAggregationLastBlockTimestamp),
        finalBlockInData: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
        parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
        finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
        l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
        aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
        lastFinalizedShnarf: generateRandomBytes(32),
      });

      await lineaRollup.setRollingHash(
        calldataAggregatedProof1To155.l1RollingHashMessageNumber,
        calldataAggregatedProof1To155.l1RollingHash,
      );

      const initialShnarf = await lineaRollup.currentFinalizedShnarf();

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocksWithProof(
          calldataAggregatedProof1To155.aggregatedProof,
          TEST_PUBLIC_VERIFIER_INDEX,
          finalizationData,
        );
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "LastFinalizedShnarfWrong", [
        initialShnarf,
        finalizationData.lastFinalizedShnarf,
      ]);
    });

    it("Should revert when proofType is invalid", async () => {
      const submissionDataBeforeFinalization = generateCallDataSubmission(0, 4);
      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: 30_000_000,
          });
        index++;
      }

      const finalizationData = await generateFinalizationData({
        l1RollingHash: calldataAggregatedProof1To155.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(calldataAggregatedProof1To155.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(calldataAggregatedProof1To155.parentAggregationLastBlockTimestamp),
        finalBlockInData: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
        parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
        finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
        l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
        aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
        shnarfData: generateParentShnarfData(index),
      });

      finalizationData.lastFinalizedShnarf = generateParentSubmissionDataForIndex(0).shnarf;

      await lineaRollup.setRollingHash(
        calldataAggregatedProof1To155.l1RollingHashMessageNumber,
        calldataAggregatedProof1To155.l1RollingHash,
      );

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocksWithProof(calldataAggregatedProof1To155.aggregatedProof, 99, finalizationData);
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
            gasLimit: 30_000_000,
          });
        index++;
      }

      const finalizationData = await generateFinalizationData({
        l1RollingHash: calldataAggregatedProof1To155.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(calldataAggregatedProof1To155.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(calldataAggregatedProof1To155.parentAggregationLastBlockTimestamp),
        finalBlockInData: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
        parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
        finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
        l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
        aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
        shnarfData: generateParentShnarfData(index),
      });

      finalizationData.lastFinalizedShnarf = generateParentSubmissionDataForIndex(0).shnarf;

      await lineaRollup.setRollingHash(
        calldataAggregatedProof1To155.l1RollingHashMessageNumber,
        calldataAggregatedProof1To155.l1RollingHash,
      );

      // removing the verifier index
      await lineaRollup.connect(securityCouncil).unsetVerifierAddress(TEST_PUBLIC_VERIFIER_INDEX);

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocksWithProof(
          calldataAggregatedProof1To155.aggregatedProof,
          TEST_PUBLIC_VERIFIER_INDEX,
          finalizationData,
        );
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
            gasLimit: 30_000_000,
          });
        index++;
      }

      const finalizationData = await generateFinalizationData({
        l1RollingHash: calldataAggregatedProof1To155.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(calldataAggregatedProof1To155.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(calldataAggregatedProof1To155.parentAggregationLastBlockTimestamp),
        finalBlockInData: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
        parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
        finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
        l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
        aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
        shnarfData: generateParentShnarfData(index),
      });

      finalizationData.lastFinalizedShnarf = generateParentSubmissionDataForIndex(0).shnarf;

      await lineaRollup.setRollingHash(
        calldataAggregatedProof1To155.l1RollingHashMessageNumber,
        calldataAggregatedProof1To155.l1RollingHash,
      );

      //     aggregatedProof1To81.aggregatedProof, // wrong proof on purpose
      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocksWithProof(aggregatedProof1To81.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
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
            gasLimit: 30_000_000,
          });
        index++;
      }

      const finalizationData = await generateFinalizationData({
        l1RollingHash: calldataAggregatedProof1To155.l1RollingHash,
        l1RollingHashMessageNumber: BigInt(calldataAggregatedProof1To155.l1RollingHashMessageNumber),
        lastFinalizedTimestamp: BigInt(calldataAggregatedProof1To155.parentAggregationLastBlockTimestamp),
        finalBlockInData: BigInt(calldataAggregatedProof1To155.finalBlockNumber),
        parentStateRootHash: calldataAggregatedProof1To155.parentStateRootHash,
        finalTimestamp: BigInt(calldataAggregatedProof1To155.finalTimestamp),
        l2MerkleRoots: calldataAggregatedProof1To155.l2MerkleRoots,
        l2MerkleTreesDepth: BigInt(calldataAggregatedProof1To155.l2MerkleTreesDepth),
        l2MessagingBlocksOffsets: calldataAggregatedProof1To155.l2MessagingBlocksOffsets,
        aggregatedProof: calldataAggregatedProof1To155.aggregatedProof,
        shnarfData: generateParentShnarfData(1),
      });

      finalizationData.lastFinalizedShnarf = generateParentSubmissionDataForIndex(1).shnarf;

      await lineaRollup.setRollingHash(
        calldataAggregatedProof1To155.l1RollingHashMessageNumber,
        calldataAggregatedProof1To155.l1RollingHash,
      );

      const finalizeCall = lineaRollup
        .connect(operator)
        .finalizeBlocksWithProof(
          calldataAggregatedProof1To155.aggregatedProof,
          TEST_PUBLIC_VERIFIER_INDEX,
          finalizationData,
        );
      await expectRevertWithCustomError(lineaRollup, finalizeCall, "LastFinalizedShnarfWrong", [
        calldataAggregatedProof1To155.parentAggregationFinalShnarf,
        finalizationData.lastFinalizedShnarf,
      ]);
    });

    it("Should successfully finalize 1-81 and then 82-153 in two separate finalizations", async () => {
      const submissionDataBeforeFinalization = generateCallDataSubmissionMultipleProofs(0, 4);
      let index = 0;
      for (const data of submissionDataBeforeFinalization) {
        const parentAndExpectedShnarf = generateParentAndExpectedShnarfForMulitpleIndex(index);
        await lineaRollup
          .connect(operator)
          .submitDataAsCalldata(data, parentAndExpectedShnarf.parentShnarf, parentAndExpectedShnarf.expectedShnarf, {
            gasLimit: 30_000_000,
          });
        index++;
      }

      await expectSuccessfulFinalizeWithProof(
        aggregatedProof1To81,
        2,
        secondCompressedDataContent.finalStateRootHash,
        generateParentShnarfData,
        true,
      );

      await expectSuccessfulFinalizeWithProof(
        aggregatedProof82To153,
        4,
        fourthMultipleCompressedDataContent.finalStateRootHash,
        generateParentShnarfData,
        true,
        aggregatedProof1To81.l1RollingHash,
        BigInt(aggregatedProof1To81.l1RollingHashMessageNumber),
      );
    });
  });

  describe("Validate L2 computed rolling hash", () => {
    it("Should revert if l1 message number == 0 and l1 rolling hash is not empty", async () => {
      const l1MessageNumber = 0;
      const l1RollingHash = generateRandomBytes(32);

      await expectRevertWithCustomError(
        lineaRollup,
        lineaRollup.validateL2ComputedRollingHash(l1MessageNumber, l1RollingHash),
        "MissingMessageNumberForRollingHash",
        [l1RollingHash],
      );
    });

    it("Should revert if l1 message number != 0 and l1 rolling hash is empty", async () => {
      const l1MessageNumber = 1n;
      const l1RollingHash = HASH_ZERO;

      await expectRevertWithCustomError(
        lineaRollup,
        lineaRollup.validateL2ComputedRollingHash(l1MessageNumber, l1RollingHash),
        "MissingRollingHashForMessageNumber",
        [l1MessageNumber],
      );
    });

    it("Should revert if l1RollingHash does not exist on L1", async () => {
      const l1MessageNumber = 1n;
      const l1RollingHash = generateRandomBytes(32);

      await expectRevertWithCustomError(
        lineaRollup,
        lineaRollup.validateL2ComputedRollingHash(l1MessageNumber, l1RollingHash),
        "L1RollingHashDoesNotExistOnL1",
        [l1MessageNumber, l1RollingHash],
      );
    });

    it("Should succeed if l1 message number == 0 and l1 rolling hash is empty", async () => {
      const l1MessageNumber = 0;
      const l1RollingHash = HASH_ZERO;
      await expect(lineaRollup.validateL2ComputedRollingHash(l1MessageNumber, l1RollingHash)).to.not.be.reverted;
    });

    it("Should succeed if l1 message number != 0, l1 rolling hash is not empty and exists on L1", async () => {
      const l1MessageNumber = 1n;
      const messageHash = generateRandomBytes(32);

      await lineaRollup.addRollingHash(l1MessageNumber, messageHash);

      const l1RollingHash = calculateRollingHash(HASH_ZERO, messageHash);

      await expect(lineaRollup.validateL2ComputedRollingHash(l1MessageNumber, l1RollingHash)).to.not.be.reverted;
    });
  });

  describe("Calculate Y value for Compressed Data", () => {
    it("Should successfully calculate y", async () => {
      const compressedDataBytes = ethers.decodeBase64(compressedData);

      expect(await lineaRollup.calculateY(compressedDataBytes, expectedX, { gasLimit: 30_000_000 })).to.equal(
        expectedY,
      );
    });

    it("Should revert if first byte is no zero", async () => {
      const compressedDataBytes = encodeData(
        ["bytes32", "bytes32", "bytes32"],
        [generateRandomBytes(32), HASH_WITHOUT_ZERO_FIRST_BYTE, generateRandomBytes(32)],
      );

      await expectRevertWithCustomError(
        lineaRollup,
        lineaRollup.calculateY(compressedDataBytes, expectedX, { gasLimit: 30_000_000 }),
        "FirstByteIsNotZero",
      );
    });

    it("Should revert if bytes length is not a multiple of 32", async () => {
      const compressedDataBytes = generateRandomBytes(56);

      await expectRevertWithCustomError(
        lineaRollup,
        lineaRollup.calculateY(compressedDataBytes, expectedX, { gasLimit: 30_000_000 }),
        "BytesLengthNotMultipleOf32",
      );
    });
  });

  async function sendBlobTransaction(startIndex: number, finalIndex: number, isMultiple: boolean = false) {
    const operatorHDSigner = getWalletForIndex(2);
    const lineaRollupAddress = await lineaRollup.getAddress();

    const {
      blobDataSubmission: blobSubmission,
      compressedBlobs: compressedBlobs,
      parentShnarf: parentShnarf,
      finalShnarf: finalShnarf,
    } = generateBlobDataSubmission(startIndex, finalIndex, isMultiple);

    const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
      blobSubmission,
      parentShnarf,
      finalShnarf,
    ]);

    const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
    const nonce = await operatorHDSigner.getNonce();

    const transaction = Transaction.from({
      data: encodedCall,
      maxPriorityFeePerGas: maxPriorityFeePerGas!,
      maxFeePerGas: maxFeePerGas!,
      to: lineaRollupAddress,
      chainId: (await ethers.provider.getNetwork()).chainId,
      type: 3,
      nonce: nonce,
      value: 0,
      gasLimit: 5_000_000,
      kzg,
      maxFeePerBlobGas: 1n,
      blobs: compressedBlobs,
    });

    const signedTx = await operatorHDSigner.signTransaction(transaction);
    await ethers.provider.broadcastTransaction(signedTx);
  }

  async function expectSuccessfulFinalizeWithProof(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    proofData: any,
    blobParentShnarfIndex: number,
    finalStateRootHash: string,
    shnarfDataGenerator: ShnarfDataGenerator,
    isMultiple: boolean = false,
    lastFinalizedRollingHash: string = HASH_ZERO,
    lastFinalizedMessageNumber: bigint = 0n,
  ) {
    const finalizationData = await generateFinalizationData({
      l1RollingHash: proofData.l1RollingHash,
      l1RollingHashMessageNumber: BigInt(proofData.l1RollingHashMessageNumber),
      lastFinalizedTimestamp: BigInt(proofData.parentAggregationLastBlockTimestamp),
      finalBlockInData: BigInt(proofData.finalBlockNumber),
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

    finalizationData.lastFinalizedShnarf = proofData.parentAggregationFinalShnarf;

    await lineaRollup.setRollingHash(proofData.l1RollingHashMessageNumber, proofData.l1RollingHash);

    const finalizeCompressedCall = lineaRollup
      .connect(operator)
      .finalizeBlocksWithProof(proofData.aggregatedProof, TEST_PUBLIC_VERIFIER_INDEX, finalizationData);
    const eventArgs = [BigInt(proofData.finalBlockNumber), finalizationData.parentStateRootHash, finalStateRootHash];

    await expectEvent(lineaRollup, finalizeCompressedCall, "BlocksVerificationDone", eventArgs);

    const [expectedFinalStateRootHash, lastFinalizedBlockNumber, lastFinalizedState] = await Promise.all([
      lineaRollup.stateRootHashes(finalizationData.finalBlockInData),
      lineaRollup.currentL2BlockNumber(),
      lineaRollup.currentFinalizedState(),
    ]);

    expect(expectedFinalStateRootHash).to.equal(finalizationData.shnarfData.finalStateRootHash);
    expect(lastFinalizedBlockNumber).to.equal(finalizationData.finalBlockInData);
    expect(lastFinalizedState).to.equal(
      generateKeccak256(
        ["uint256", "bytes32", "uint256"],
        [finalizationData.l1RollingHashMessageNumber, finalizationData.l1RollingHash, finalizationData.finalTimestamp],
      ),
    );
  }

  describe("LineaRollup Upgradeable Tests", () => {
    async function deployLineaRollupFixture() {
      const PlonkVerifierFactory = await ethers.getContractFactory("TestPlonkVerifierForDataAggregation");
      const plonkVerifier = await PlonkVerifierFactory.deploy();
      await plonkVerifier.waitForDeployment();

      verifier = await plonkVerifier.getAddress();

      const lineaRollup = (await deployUpgradableFromFactory(
        "contracts/test-contracts/LineaRollupFlattened.sol:LineaRollupFlattened",
        [
          parentStateRootHash,
          0,
          verifier,
          securityCouncil.address,
          [operator.address],
          ONE_DAY_IN_SECONDS,
          INITIAL_WITHDRAW_LIMIT,
          1683325137n,
        ],
        {
          initializer: "initialize(bytes32,uint256,address,address,address[],uint256,uint256,uint256)",
          unsafeAllow: ["constructor"],
        },
      )) as unknown as TestLineaRollup;

      return lineaRollup;
    }

    beforeEach(async () => {
      lineaRollup = await loadFixture(deployLineaRollupFixture);
    });

    it("Should deploy and upgrade the LineaRollup contract", async () => {
      expect(await lineaRollup.currentL2BlockNumber()).to.equal(0);

      // Deploy new implementation
      const NewLineaRollupFactory = await ethers.getContractFactory("contracts/LineaRollup.sol:LineaRollup");
      const newLineaRollup = await upgrades.upgradeProxy(lineaRollup, NewLineaRollupFactory);

      await newLineaRollup.reinitializePauseTypesAndPermissions(
        [
          { addressWithRole: securityCouncil.address, role: DEFAULT_ADMIN_ROLE },
          { addressWithRole: securityCouncil.address, role: VERIFIER_SETTER_ROLE },
        ],
        pauseTypeRoles,
        unpauseTypeRoles,
        multiCallAddress,
      );

      expect(await newLineaRollup.currentL2BlockNumber()).to.equal(0);
    });

    it("Should revert with ZeroAddressNotAllowed when addressWithRole is zero address in reinitializePauseTypesAndPermissions", async () => {
      // Deploy new implementation
      const NewLineaRollupFactory = await ethers.getContractFactory("contracts/LineaRollup.sol:LineaRollup");
      const newLineaRollup = await upgrades.upgradeProxy(lineaRollup, NewLineaRollupFactory);

      const roleAddresses = [
        { addressWithRole: ZeroAddress, role: DEFAULT_ADMIN_ROLE },
        { addressWithRole: securityCouncil.address, role: VERIFIER_SETTER_ROLE },
        { addressWithRole: operator.address, role: OPERATOR_ROLE },
      ];

      await expectRevertWithCustomError(
        newLineaRollup,
        newLineaRollup.reinitializePauseTypesAndPermissions(
          roleAddresses,
          pauseTypeRoles,
          unpauseTypeRoles,
          multiCallAddress,
        ),
        "ZeroAddressNotAllowed",
      );
    });
  });
});
