import { describe, expect, it } from "@jest/globals";
import { config } from "./config/tests-config";

const l2AccountManager = config.getL2AccountManager();

describe("Gas limit test suite", () => {
  const l2Provider = config.getL2Provider();

  it.concurrent("Should successfully invoke GasLimitTestContract.setGasLimit()", async () => {
    const account = await l2AccountManager.generateAccount();
    const gasLimitTestContract = config.getGasLimitTestContract(account);
    const nonce = await l2Provider.getTransactionCount(account.address, "pending");
    const { maxPriorityFeePerGas, maxFeePerGas } = await l2Provider.getFeeData();

    const tx = await gasLimitTestContract.connect(account).setGasLimit({
      nonce: nonce,
      maxPriorityFeePerGas: maxPriorityFeePerGas,
      maxFeePerGas: maxFeePerGas,
    });

    const receipt = await tx.wait();
    expect(receipt?.status).toEqual(1);
  });
});
