import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ForcedTransactionGateway, Mimc, TestLineaRollup, TestEip1559RlpEncoder } from "contracts/typechain-types";
import transactionWithoutCalldata from "../../_testData/eip1559RlpEncoderTransactions/withoutCalldata.json";
import transactionWithLargeCalldata from "../../_testData/eip1559RlpEncoderTransactions/withLargeCalldata.json";
import transactionWithCalldataAndAccessList from "../../_testData/eip1559RlpEncoderTransactions/withCalldataAndAccessList.json";
import transactionWithCalldata from "../../_testData/eip1559RlpEncoderTransactions/withCalldata.json";
import l2SendMessageTransaction from "../../_testData/eip1559RlpEncoderTransactions/l2SendMessage.json";
import { ethers } from "hardhat";

import {
  getAccountsFixture,
  deployForcedTransactionGatewayFixture,
  setNextExpectedL2BlockNumberForForcedTx,
  getForcedTransactionRollingHash,
} from "./../helpers";
import {
  buildAccessErrorMessage,
  buildEip1559Transaction,
  expectEvent,
  expectRevertWithCustomError,
  expectRevertWithReason,
  generateKeccak256,
  generateRandomBytes,
} from "../../common/helpers";
import {
  ADDRESS_ZERO,
  DEFAULT_LAST_FINALIZED_TIMESTAMP,
  FORCED_TRANSACTION_SENDER_ROLE,
  HASH_ZERO,
  LINEA_MAINNET_CHAIN_ID,
  MAX_GAS_LIMIT,
  MAX_INPUT_LENGTH_LIMIT,
  THREE_DAYS_IN_SECONDS,
  DEFAULT_FUTURE_NEXT_NETWORK_TIMESTAMP,
} from "../../common/constants";
import { toBeHex, zeroPadValue } from "ethers";
import { expect } from "chai";
import { deployFromFactory } from "../../common/deployment";

describe("Linea Rollup contract: Forced Transactions", () => {
  let lineaRollup: TestLineaRollup;
  let forcedTransactionGateway: ForcedTransactionGateway;
  let mimcLibrary: Mimc;
  let mimcLibraryAddress: string;
  let eip1559RlpEncoder: TestEip1559RlpEncoder;

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let securityCouncil: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;
  let defaultFinalizedState = {
    messageNumber: 0n,
    messageRollingHash: HASH_ZERO,
    forcedTransactionNumber: 0n,
    forcedTransactionRollingHash: HASH_ZERO,
    timestamp: DEFAULT_LAST_FINALIZED_TIMESTAMP,
  };

  before(async () => {
    ({ nonAuthorizedAccount, securityCouncil } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ lineaRollup, forcedTransactionGateway } = await loadFixture(deployForcedTransactionGatewayFixture));

    await lineaRollup
      .connect(securityCouncil)
      .grantRole(FORCED_TRANSACTION_SENDER_ROLE, await forcedTransactionGateway.getAddress());

    await lineaRollup
      .connect(securityCouncil)
      .grantRole(FORCED_TRANSACTION_SENDER_ROLE, await securityCouncil.getAddress());

    defaultFinalizedState = {
      messageNumber: 0n,
      messageRollingHash: HASH_ZERO,
      forcedTransactionNumber: 0n,
      forcedTransactionRollingHash: HASH_ZERO,
      timestamp: DEFAULT_LAST_FINALIZED_TIMESTAMP,
    };

    async function deployMimc() {
      mimcLibrary = (await deployFromFactory("Mimc")) as Mimc;
      await mimcLibrary.waitForDeployment();
      mimcLibraryAddress = await mimcLibrary.getAddress();
    }

    async function deployTestEip1559RlpEncoderFixture() {
      return deployFromFactory("TestEip1559RlpEncoder", LINEA_MAINNET_CHAIN_ID);
    }

    // Unsure why the following two lines do not work in before block
    // If we deploy mimic or eip1559RlpEncoder in before block, and try to invoke the contracts, we get a weird error
    await loadFixture(deployMimc);
    eip1559RlpEncoder = (await loadFixture(deployTestEip1559RlpEncoderFixture)) as TestEip1559RlpEncoder;
  });

  describe("Contract Construction", () => {
    it("Should fail if the Linea rollup is set as address(0)", async () => {
      const forcedTransactionGatewayFactory = await ethers.getContractFactory("ForcedTransactionGateway", {
        libraries: { Mimc: mimcLibraryAddress },
      });

      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGatewayFactory.deploy(
          ADDRESS_ZERO,
          LINEA_MAINNET_CHAIN_ID,
          THREE_DAYS_IN_SECONDS,
          MAX_GAS_LIMIT,
          MAX_INPUT_LENGTH_LIMIT,
        ),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should fail if the chainId is set to zero", async () => {
      const forcedTransactionGatewayFactory = await ethers.getContractFactory("ForcedTransactionGateway", {
        libraries: { Mimc: mimcLibraryAddress },
      });

      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGatewayFactory.deploy(
          await lineaRollup.getAddress(),
          0,
          THREE_DAYS_IN_SECONDS,
          MAX_GAS_LIMIT,
          MAX_INPUT_LENGTH_LIMIT,
        ),
        "ZeroValueNotAllowed",
      );
    });

    it("Should fail if the block buffer is set to zero", async () => {
      const forcedTransactionGatewayFactory = await ethers.getContractFactory("ForcedTransactionGateway", {
        libraries: { Mimc: mimcLibraryAddress },
      });

      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGatewayFactory.deploy(
          await lineaRollup.getAddress(),
          LINEA_MAINNET_CHAIN_ID,
          0,
          MAX_GAS_LIMIT,
          MAX_INPUT_LENGTH_LIMIT,
        ),
        "ZeroValueNotAllowed",
      );
    });

    it("Should fail if the max gas limit is set to zero", async () => {
      const forcedTransactionGatewayFactory = await ethers.getContractFactory("ForcedTransactionGateway", {
        libraries: { Mimc: mimcLibraryAddress },
      });

      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGatewayFactory.deploy(
          await lineaRollup.getAddress(),
          LINEA_MAINNET_CHAIN_ID,
          THREE_DAYS_IN_SECONDS,
          0,
          MAX_INPUT_LENGTH_LIMIT,
        ),
        "ZeroValueNotAllowed",
      );
    });

    it("Should fail if the max input limit is set to zero", async () => {
      const forcedTransactionGatewayFactory = await ethers.getContractFactory("ForcedTransactionGateway", {
        libraries: { Mimc: mimcLibraryAddress },
      });

      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGatewayFactory.deploy(
          await lineaRollup.getAddress(),
          LINEA_MAINNET_CHAIN_ID,
          THREE_DAYS_IN_SECONDS,
          MAX_GAS_LIMIT,
          0,
        ),
        "ZeroValueNotAllowed",
      );
    });
  });

  describe("Adding forced transactions", () => {
    it("Should fail if the gas limit is too high", async () => {
      const sendCall = forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(transactionWithLargeCalldata.result),
        defaultFinalizedState,
      );
      await expectRevertWithCustomError(forcedTransactionGateway, sendCall, "MaxGasLimitExceeded");
    });

    it("Should fail if the calldata input is too long", async () => {
      const forcedTransaction = buildEip1559Transaction(l2SendMessageTransaction.result);
      forcedTransaction.input = generateRandomBytes(Number(MAX_INPUT_LENGTH_LIMIT) + 1);
      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGateway.submitForcedTransaction(forcedTransaction, defaultFinalizedState),
        "CalldataInputLengthLimitExceeded",
      );
    });

    it("Should fail if the maxPriorityFeePerGas is zero", async () => {
      const forcedTransaction = buildEip1559Transaction(l2SendMessageTransaction.result);
      forcedTransaction.maxPriorityFeePerGas = 0n;
      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGateway.submitForcedTransaction(forcedTransaction, defaultFinalizedState),
        "GasFeeParametersContainZero",
        [forcedTransaction.maxFeePerGas, 0],
      );
    });

    it("Should fail if the maxFeePerGas is zero", async () => {
      const forcedTransaction = buildEip1559Transaction(l2SendMessageTransaction.result);
      forcedTransaction.maxFeePerGas = 0n;
      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGateway.submitForcedTransaction(forcedTransaction, defaultFinalizedState),
        "GasFeeParametersContainZero",
        [0, forcedTransaction.maxPriorityFeePerGas],
      );
    });

    it("Should fail if maxPriorityFeePerGas > maxFeePerGas", async () => {
      const forcedTransaction = buildEip1559Transaction(l2SendMessageTransaction.result);
      forcedTransaction.maxPriorityFeePerGas = 2n;
      forcedTransaction.maxFeePerGas = 1n;
      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGateway.submitForcedTransaction(forcedTransaction, defaultFinalizedState),
        "MaxPriorityFeePerGasHigherThanMaxFee",
        [forcedTransaction.maxFeePerGas, forcedTransaction.maxPriorityFeePerGas],
      );
    });

    it("Should fail if YParity > 1", async () => {
      const forcedTransaction = buildEip1559Transaction(l2SendMessageTransaction.result);
      forcedTransaction.yParity = 2n;
      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGateway.submitForcedTransaction(forcedTransaction, defaultFinalizedState),
        "YParityGreaterThanOne",
        [forcedTransaction.yParity],
      );
    });

    it("Should fail if the To address is less than 21", async () => {
      const forcedTransaction = buildEip1559Transaction(l2SendMessageTransaction.result);
      for (let i = 0; i < 21; i++) {
        forcedTransaction.to = ethers.getAddress(zeroPadValue(toBeHex(i), 20));
        await expectRevertWithCustomError(
          forcedTransactionGateway,
          forcedTransactionGateway.submitForcedTransaction(forcedTransaction, defaultFinalizedState),
          "ToAddressTooLow",
          [],
        );
      }
    });

    it("Should fail if the last finalized state hash does not match", async () => {
      const forcedTransaction = buildEip1559Transaction(l2SendMessageTransaction.result);

      const defaultFinalizedStateHash = generateKeccak256(
        ["uint256", "bytes32", "uint256", "bytes32", "uint256"],
        [
          defaultFinalizedState.messageNumber,
          defaultFinalizedState.messageRollingHash,
          defaultFinalizedState.forcedTransactionNumber,
          defaultFinalizedState.forcedTransactionRollingHash,
          defaultFinalizedState.timestamp,
        ],
      );

      const corruptedFinalizedStateStruct = {
        ...defaultFinalizedState,
        forcedTransactionNumber: 1n,
      };

      const corruptedFinalizationStateHash = generateKeccak256(
        ["uint256", "bytes32", "uint256", "bytes32", "uint256"],
        [
          corruptedFinalizedStateStruct.messageNumber,
          corruptedFinalizedStateStruct.messageRollingHash,
          corruptedFinalizedStateStruct.forcedTransactionNumber,
          corruptedFinalizedStateStruct.forcedTransactionRollingHash,
          corruptedFinalizedStateStruct.timestamp,
        ],
      );

      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGateway.submitForcedTransaction(forcedTransaction, corruptedFinalizedStateStruct),
        "FinalizationStateIncorrect",
        [defaultFinalizedStateHash, corruptedFinalizationStateHash],
      );
    });

    it("Should fail LineaRollup.storeForcedTransaction if not FORCED_TRANSACTION_SENDER_ROLE", async () => {
      await expectRevertWithReason(
        lineaRollup.connect(nonAuthorizedAccount).storeForcedTransaction(99n, 121n, generateRandomBytes(32)),
        buildAccessErrorMessage(nonAuthorizedAccount, FORCED_TRANSACTION_SENDER_ROLE),
      );
    });

    it("Should fail if the second transaction is expected on the same block", async () => {
      const expectedBlockNumber = await setNextExpectedL2BlockNumberForForcedTx(
        lineaRollup,
        DEFAULT_FUTURE_NEXT_NETWORK_TIMESTAMP,
        defaultFinalizedState.timestamp,
      );

      await forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(l2SendMessageTransaction.result),
        defaultFinalizedState,
      );

      await expectRevertWithCustomError(
        lineaRollup,
        lineaRollup.connect(securityCouncil).storeForcedTransaction(2, expectedBlockNumber, generateRandomBytes(32)),
        "ForcedTransactionExistsForBlock",
        [expectedBlockNumber],
      );
    });

    it("Should fail if the second transaction is expected on the same transaction number", async () => {
      const expectedBlockNumber = await setNextExpectedL2BlockNumberForForcedTx(
        lineaRollup,
        DEFAULT_FUTURE_NEXT_NETWORK_TIMESTAMP,
        defaultFinalizedState.timestamp,
      );

      await forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(l2SendMessageTransaction.result),
        defaultFinalizedState,
      );

      await expectRevertWithCustomError(
        lineaRollup,
        lineaRollup.connect(securityCouncil).storeForcedTransaction(1, expectedBlockNumber, generateRandomBytes(32)),
        "ForcedTransactionExistsForTransactionNumber",
        [1],
      );
    });

    it("Should submit the forced transaction with no calldata", async () => {
      await forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(transactionWithoutCalldata.result),
        defaultFinalizedState,
      );
    });

    it("Should submit the forced transaction with calldata and access list", async () => {
      await forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(transactionWithCalldataAndAccessList.result),
        defaultFinalizedState,
      );
    });

    it("Should submit the forced transaction with calldata", async () => {
      await forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(transactionWithCalldata.result),
        defaultFinalizedState,
      );
    });

    it("Should submit the forced L2 SendMessage transaction with calldata", async () => {
      await forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(l2SendMessageTransaction.result),
        defaultFinalizedState,
      );
    });

    it("Should emit the ForcedTransactionAdded event on adding a transaction", async () => {
      // use a way future dated timestamp and mimc the calculation for the block number
      const expectedBlockNumber = await setNextExpectedL2BlockNumberForForcedTx(
        lineaRollup,
        1954213624n,
        defaultFinalizedState.timestamp,
      );

      const expectedForcedTransactionNumber = 1n;

      const expectedMimcHashWithPreviousZeroValueRollingHash = await getForcedTransactionRollingHash(
        mimcLibrary,
        lineaRollup,
        eip1559RlpEncoder,
        buildEip1559Transaction(l2SendMessageTransaction.result),
        expectedBlockNumber,
        l2SendMessageTransaction?.result?.from,
      );

      const expectedEventArgs = [
        expectedForcedTransactionNumber,
        ethers.getAddress(l2SendMessageTransaction.result.from),
        expectedBlockNumber,
        expectedMimcHashWithPreviousZeroValueRollingHash,
        l2SendMessageTransaction.rlpEncodedUnsigned,
        l2SendMessageTransaction.rlpEncodedSigned,
      ];

      await expectEvent(
        forcedTransactionGateway,
        forcedTransactionGateway.submitForcedTransaction(
          buildEip1559Transaction(l2SendMessageTransaction.result),
          defaultFinalizedState,
        ),
        "ForcedTransactionAdded",
        expectedEventArgs,
      );
    });

    it("Should change rolling hash with different expected block number", async () => {
      // use a way future dated timestamp and mimic the calculation for the block number
      const expectedBlockNumber = await setNextExpectedL2BlockNumberForForcedTx(
        lineaRollup,
        1754213624n,
        defaultFinalizedState.timestamp,
      );

      const expectedForcedTransactionNumber = 1n;

      const expectedMimcHashWithPreviousZeroValueRollingHash = await getForcedTransactionRollingHash(
        mimcLibrary,
        lineaRollup,
        eip1559RlpEncoder,
        buildEip1559Transaction(l2SendMessageTransaction.result),
        expectedBlockNumber,
        l2SendMessageTransaction?.result?.from,
      );

      const expectedEventArgs = [
        expectedForcedTransactionNumber,
        ethers.getAddress(l2SendMessageTransaction.result.from),
        expectedBlockNumber,
        expectedMimcHashWithPreviousZeroValueRollingHash,
        l2SendMessageTransaction.rlpEncodedUnsigned,
        l2SendMessageTransaction.rlpEncodedSigned,
      ];

      await expectEvent(
        forcedTransactionGateway,
        forcedTransactionGateway.submitForcedTransaction(
          buildEip1559Transaction(l2SendMessageTransaction.result),
          defaultFinalizedState,
        ),
        "ForcedTransactionAdded",
        expectedEventArgs,
      );
    });

    it("Updates the next message number on the Linea Rollup", async () => {
      expect(await lineaRollup.nextForcedTransactionNumber()).equal(1);

      await forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(l2SendMessageTransaction.result),
        defaultFinalizedState,
      );

      expect(await lineaRollup.nextForcedTransactionNumber()).equal(2);
    });

    it("Updates the forcedTransactionL2BlockNumbers on the Linea Rollup", async () => {
      expect(await lineaRollup.forcedTransactionL2BlockNumbers(1)).equal(0);

      await forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(l2SendMessageTransaction.result),
        defaultFinalizedState,
      );

      expect(await lineaRollup.forcedTransactionL2BlockNumbers(1)).greaterThan(0);
    });

    it("Updates the forcedTransactionRollingHashes on the Linea Rollup", async () => {
      expect(await lineaRollup.forcedTransactionRollingHashes(1)).equal(HASH_ZERO);
      const expectedBlockNumber = await setNextExpectedL2BlockNumberForForcedTx(
        lineaRollup,
        DEFAULT_FUTURE_NEXT_NETWORK_TIMESTAMP,
        defaultFinalizedState.timestamp,
      );
      const expectedForcedTxRollingHash = await getForcedTransactionRollingHash(
        mimcLibrary,
        lineaRollup,
        eip1559RlpEncoder,
        buildEip1559Transaction(l2SendMessageTransaction.result),
        expectedBlockNumber,
        l2SendMessageTransaction?.result?.from,
      );

      await forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(l2SendMessageTransaction.result),
        defaultFinalizedState,
      );

      expect(await lineaRollup.forcedTransactionRollingHashes(1)).equal(expectedForcedTxRollingHash);
    });
  });
});
