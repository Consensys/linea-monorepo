import { describe, expect, it } from "@jest/globals";
import { pollForContractMethodReturnValueExceedTarget } from "./common/utils";
import { config } from "./config/tests-config";
import { ContractTransactionReceipt, Wallet } from "ethers";

const l2AccountManager = config.getL2AccountManager();
const l2Provider = config.getL2Provider();

describe("Gas limit test suite", () => {
  const setGasLimit = async (account: Wallet): Promise<ContractTransactionReceipt | null> => {
    const opcodeTestContract = config.getOpcodeTestContract(account);
    const nonce = await l2Provider.getTransactionCount(account.address, "pending");
    const { maxPriorityFeePerGas, maxFeePerGas } = await l2Provider.getFeeData();

    const tx = await opcodeTestContract.connect(account).setGasLimit({
      nonce: nonce,
      maxPriorityFeePerGas: maxPriorityFeePerGas,
      maxFeePerGas: maxFeePerGas,
    });

    const receipt = await tx.wait();
    return receipt;
  };

  it.concurrent("Should successfully invoke OpcodeTestContract.setGasLimit()", async () => {
    const account = await l2AccountManager.generateAccount();
    const receipt = await setGasLimit(account);
    expect(receipt?.status).toEqual(1);
  });

  it("Should successfully finalize OpcodeTestContract.setGasLimit()", async () => {
    const account = await l2AccountManager.generateAccount();
    const lineaRollupV6 = config.getLineaRollupContract();

    const txReceipt = await setGasLimit(account);
    expect(txReceipt?.status).toEqual(1);
    // Ok to type assertion here, because txReceipt won't be null if it passed above assertion.
    const txBlockNumber = <number>txReceipt?.blockNumber;

    console.log(`Waiting for ${txBlockNumber} to be finalized...`);

    const isBlockFinalized = await pollForContractMethodReturnValueExceedTarget(
      lineaRollupV6.currentL2BlockNumber,
      BigInt(txBlockNumber),
    );

    expect(isBlockFinalized).toEqual(true);
  });
});
