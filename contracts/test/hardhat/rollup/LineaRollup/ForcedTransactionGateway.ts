import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ForcedTransactionGateway, TestLineaRollup } from "contracts/typechain-types";
import transactionWithoutCalldata from "../../_testData/eip1559RlpEncoderTransactions/withoutCalldata.json";
import transactionWithLargeCalldata from "../../_testData/eip1559RlpEncoderTransactions/withLargeCalldata.json";
import transactionWithCalldataAndAccessList from "../../_testData/eip1559RlpEncoderTransactions/withCalldataAndAccessList.json";
import transactionWithCalldata from "../../_testData/eip1559RlpEncoderTransactions/withCalldata.json";
import l2SendMessageTransaction from "../../_testData/eip1559RlpEncoderTransactions/l2SendMessage.json";

import { getAccountsFixture, deployForcedTransactionGatewayFixture } from "./../helpers";
import { buildEip1559Transaction, expectRevertWithCustomError } from "../../common/helpers";
import { DEFAULT_LAST_FINALIZED_TIMESTAMP, FORCED_TRANSACTION_SENDER_ROLE, HASH_ZERO } from "../../common/constants";

describe("Linea Rollup contract: Forced Transactions", () => {
  let lineaRollup: TestLineaRollup;
  let forcedTransactionGateway: ForcedTransactionGateway;

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let securityCouncil: SignerWithAddress;

  const defaultFinalizedState = {
    messageNumber: 0n,
    messageRollingHash: HASH_ZERO,
    forcedTransactionNumber: 0n,
    forcedTransactionRollingHash: HASH_ZERO,
    timestamp: DEFAULT_LAST_FINALIZED_TIMESTAMP,
  };

  before(async () => {
    ({ securityCouncil } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ lineaRollup, forcedTransactionGateway } = await loadFixture(deployForcedTransactionGatewayFixture));

    await lineaRollup
      .connect(securityCouncil)
      .grantRole(FORCED_TRANSACTION_SENDER_ROLE, await forcedTransactionGateway.getAddress());
  });

  describe("Adding forced transactions", () => {
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

    it("Should fail if the gas limit is too high", async () => {
      const sendCall = forcedTransactionGateway.submitForcedTransaction(
        buildEip1559Transaction(transactionWithLargeCalldata.result),
        defaultFinalizedState,
      );
      await expectRevertWithCustomError(forcedTransactionGateway, sendCall, "MaxGasLimitExceeded");
    });
  });
});
