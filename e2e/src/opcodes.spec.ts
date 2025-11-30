import { describe, expect, it } from "@jest/globals";
import { L2RpcEndpoint } from "./config/tests-config/setup/clients/l2-client";
import { config } from "./config/tests-config/setup";
import { encodeFunctionCall, estimateLineaGas } from "./common/utils";
import { OpcodeTesterAbi } from "./generated";

const l2AccountManager = config.getL2AccountManager();

describe("Opcodes test suite", () => {
  const l2PublicClient = config.l2PublicClient({ type: L2RpcEndpoint.BesuNode });

  it.concurrent("Should be able to estimate the opcode execution gas using linea_estimateGas endpoint", async () => {
    const account = await l2AccountManager.generateAccount();
    const opcodeTester = config.l2PublicClient().getOpcodeTesterContract();

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
    const l2PublicClient = config.l2PublicClient();
    const opcodeTesterRead = l2PublicClient.getOpcodeTesterContract();
    const opcodeTesterWrite = config.l2WalletClient({ account }).getOpcodeTesterContract();

    const valueBeforeExecution = await opcodeTesterRead.read.rollingBlockDetailComputations();

    const { maxPriorityFeePerGas, maxFeePerGas, gasLimit } = await estimateLineaGas(l2PublicClient, {
      account,
      to: opcodeTesterRead.address,
      data: encodeFunctionCall({
        abi: OpcodeTesterAbi,
        functionName: "executeAllOpcodes",
      }),
    });

    logger.debug(
      `Fetched fee data for opcode execution. maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas} gasLimit=${gasLimit}`,
    );

    const txHash = await opcodeTesterWrite.write.executeAllOpcodes({
      gas: gasLimit,
      maxFeePerGas,
      maxPriorityFeePerGas,
    });

    await l2PublicClient.waitForTransactionReceipt({ hash: txHash, timeout: 20_000 });

    const valueAfterExecution = await opcodeTesterRead.read.rollingBlockDetailComputations();

    logger.debug(`Value before execution: ${valueBeforeExecution}, value after execution: ${valueAfterExecution}`);
    expect(valueBeforeExecution).not.toEqual(valueAfterExecution);

    logger.debug("All opcodes executed successfully");
  });
});
