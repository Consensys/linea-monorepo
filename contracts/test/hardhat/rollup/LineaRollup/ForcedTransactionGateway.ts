import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture, time as networkTime } from "@nomicfoundation/hardhat-network-helpers";
import { ForcedTransactionGateway, Mimc, TestLineaRollup } from "contracts/typechain-types";
import transactionWithoutCalldata from "../../_testData/eip1559RlpEncoderTransactions/withoutCalldata.json";
import transactionWithLargeCalldata from "../../_testData/eip1559RlpEncoderTransactions/withLargeCalldata.json";
import transactionWithCalldataAndAccessList from "../../_testData/eip1559RlpEncoderTransactions/withCalldataAndAccessList.json";
import transactionWithCalldata from "../../_testData/eip1559RlpEncoderTransactions/withCalldata.json";
import l2SendMessageTransaction from "../../_testData/eip1559RlpEncoderTransactions/l2SendMessage.json";
import { ethers } from "hardhat";

import { getAccountsFixture, deployForcedTransactionGatewayFixture } from "./../helpers";
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
} from "../../common/constants";
import { toBeHex, zeroPadValue } from "ethers";
import { expect } from "chai";
import { deployFromFactory } from "../../common/deployment";

describe("Linea Rollup contract: Forced Transactions", () => {
  let lineaRollup: TestLineaRollup;
  let forcedTransactionGateway: ForcedTransactionGateway;
  let mimcLibraryAddress: string;

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
    const mimcLibrary = (await deployFromFactory("Mimc")) as unknown as Mimc;
    await mimcLibrary.waitForDeployment();
    mimcLibraryAddress = await mimcLibrary.getAddress();
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

    it("Should fail if the maxFeePerGas is zero", async () => {
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

    it("Should fail if YParity Greater Than One", async () => {
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

    it("Should fail if the last finalized state does not match", async () => {
      const forcedTransaction = buildEip1559Transaction(l2SendMessageTransaction.result);

      const realFinalizedState = generateKeccak256(
        ["uint256", "bytes32", "uint256", "bytes32", "uint256"],
        [
          defaultFinalizedState.messageNumber,
          defaultFinalizedState.messageRollingHash,
          defaultFinalizedState.forcedTransactionNumber,
          defaultFinalizedState.forcedTransactionRollingHash,
          defaultFinalizedState.timestamp,
        ],
      );

      defaultFinalizedState.forcedTransactionNumber = 1n;

      const failedFinalizationState = generateKeccak256(
        ["uint256", "bytes32", "uint256", "bytes32", "uint256"],
        [
          defaultFinalizedState.messageNumber,
          defaultFinalizedState.messageRollingHash,
          defaultFinalizedState.forcedTransactionNumber,
          defaultFinalizedState.forcedTransactionRollingHash,
          defaultFinalizedState.timestamp,
        ],
      );

      await expectRevertWithCustomError(
        forcedTransactionGateway,
        forcedTransactionGateway.submitForcedTransaction(forcedTransaction, defaultFinalizedState),
        "FinalizationStateIncorrect",
        [realFinalizedState, failedFinalizationState],
      );
    });

    it("Should fail to call store on the Linea rollup if not the gateway", async () => {
      await expectRevertWithReason(
        lineaRollup.connect(nonAuthorizedAccount).storeForcedTransaction(99n, 121n, generateRandomBytes(32)),
        buildAccessErrorMessage(nonAuthorizedAccount, FORCED_TRANSACTION_SENDER_ROLE),
      );
    });

    it("Should fail if the second transaction is expected on the same block", async () => {
      await networkTime.setNextBlockTimestamp(1954213624);
      const lastFinalizedBlock = await lineaRollup.currentL2BlockNumber();

      const expectedBlockNumber =
        1954213624n - defaultFinalizedState.timestamp + lastFinalizedBlock + BigInt(THREE_DAYS_IN_SECONDS);

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
      await networkTime.setNextBlockTimestamp(1974213624n);
      const lastFinalizedBlock = await lineaRollup.currentL2BlockNumber();

      const expectedBlockNumber =
        1974213624n - defaultFinalizedState.timestamp + lastFinalizedBlock + BigInt(THREE_DAYS_IN_SECONDS);

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
      // use a way future dated timestamp and mimic the calculation for the block number
      await networkTime.setNextBlockTimestamp(1954213624);
      const expectedForcedTransactionNumber = 1n;

      // TODO: Manually compute with Mimc for more dynamic testing
      const expectedMimcHashWithPreviousZeroValueRollingHash =
        "0x06f999b87d23e5d8b579a906300ca23b8029080d071517b75774b2e6b9abda8c";
      const lastFinalizedBlock = await lineaRollup.currentL2BlockNumber();
      const expectedBlockNumber =
        1954213624n - defaultFinalizedState.timestamp + lastFinalizedBlock + BigInt(THREE_DAYS_IN_SECONDS);

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
      await networkTime.setNextBlockTimestamp(1754213624);
      const expectedForcedTransactionNumber = 1n;

      // TODO: Manually compute with Mimc for more dynamic testing
      const expectedMimcHashWithPreviousZeroValueRollingHash =
        "0x105d807fa47dc19c25ca7ea12d66f8eea2428a916ae8db7e458b9a58b1ef6041";
      const lastFinalizedBlock = await lineaRollup.currentL2BlockNumber();
      const expectedBlockNumber =
        1754213624n - defaultFinalizedState.timestamp + lastFinalizedBlock + BigInt(THREE_DAYS_IN_SECONDS);

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

      await forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(l2SendMessageTransaction.result),
        defaultFinalizedState,
      );

      expect(await lineaRollup.forcedTransactionRollingHashes(1)).not.equal(HASH_ZERO);
    });
  });
});
