import { describe, expect, it } from "@jest/globals";
import { L2RpcEndpointType } from "./config/tests-config/setup/clients/l2-client";
import { config } from "./config/tests-config/setup";
import { encodeFunctionCall, estimateLineaGas } from "./common/utils";
import { OpcodeTesterAbi } from "./generated";

const l2AccountManager = config.getL2AccountManager();

describe("Opcodes test suite", () => {
  const l2PublicClient = config.l2PublicClient({ type: L2RpcEndpointType.BesuNode });

  it.concurrent("Should be able to estimate the opcode execution gas using linea_estimateGas endpoint", async () => {
    const account = await l2AccountManager.generateAccount();
    const opcodeTester = config.l2PublicClient().contracts.opcodeTester;

    const { maxPriorityFeePerGas, maxFeePerGas, gasLimit } = await estimateLineaGas(l2PublicClient, {
      account,
      to: opcodeTester.address,
      data: encodeFunctionCall({
        abi: OpcodeTesterAbi,
        functionName: "executeAllOpcodes",
      }),
    });
    logger.debug(
      `Fetched fee data. maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas} gasLimit=${gasLimit}`,
    );

    expect(maxPriorityFeePerGas).toBeGreaterThan(0n);
    expect(maxFeePerGas).toBeGreaterThan(0n);
    expect(gasLimit).toBeGreaterThan(0n);
  });

  it.concurrent("Should be able to execute all opcodes", async () => {
    const account = await l2AccountManager.generateAccount();
    const opcodeTesterRead = config.l2PublicClient().contracts.opcodeTester;
    const opcodeTesterWrite = config.l2WalletClient({ account }).contracts.opcodeTester;

    const valueBeforeExecution = await opcodeTesterRead.rollingBlockDetailComputations();
    const txHash = await opcodeTesterWrite.executeAllOpcodes({ gas: 5_000_000n });
    await config.l2PublicClient().waitForTransactionReceipt({ hash: txHash, timeout: 20_000 });

    const valueAfterExecution = await opcodeTesterRead.rollingBlockDetailComputations();

    logger.debug(`Value before execution: ${valueBeforeExecution}, value after execution: ${valueAfterExecution}`);
    expect(valueBeforeExecution).not.toEqual(valueAfterExecution);

    logger.debug("All opcodes executed successfully");
  });
});
