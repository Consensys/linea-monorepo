import { describe, expect, it } from "@jest/globals";
import { config } from "./config/tests-config";
import { LineaEstimateGasClient } from "./common/utils";

const l2AccountManager = config.getL2AccountManager();

describe("Opcodes test suite", () => {
  const lineaEstimateGasClient = new LineaEstimateGasClient(config.getL2BesuNodeEndpoint()!);

  it.concurrent("Should be able to estimate the opcode execution using linea_estimateGas endpoint", async () => {
    const account = await l2AccountManager.generateAccount();
    const opcodeTester = config.getL2OpcodeTester(account);

    if (!opcodeTester) {
      if (config.isLocalL2Config(config.config.L2)) {
        throw new Error("Opcode tester contract address must be defined for local L2 config");
      }
      logger.info("No opcode tester contract deployed, skipping test");
      return;
    }

    const { maxPriorityFeePerGas, maxFeePerGas, gasLimit } = await lineaEstimateGasClient.lineaEstimateGas(
      account.address,
      await opcodeTester.getAddress(),
      opcodeTester.interface.encodeFunctionData("executeAllOpcodes"),
    );
    logger.debug(
      `Fetched fee data. maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas} gasLimit=${gasLimit}`,
    );

    expect(maxPriorityFeePerGas).toBeGreaterThan(0n);
    expect(maxFeePerGas).toBeGreaterThan(0n);
    expect(gasLimit).toBeGreaterThan(0n);
  });

  it.concurrent("Should be able to execute all opcodes", async () => {
    const account = await l2AccountManager.generateAccount();
    const opcodeTester = config.getL2OpcodeTester(account);

    if (!opcodeTester) {
      if (config.isLocalL2Config(config.config.L2)) {
        throw new Error("Opcode tester contract address must be defined for local L2 config");
      }
      logger.info("No opcode tester contract deployed, skipping test");
      return;
    }

    const valueBeforeExecution = await opcodeTester.rollingBlockDetailComputations();
    const executeTx = await opcodeTester.executeAllOpcodes({ gasLimit: 5_000_000 });
    await executeTx.wait(1, 20_000);

    const valueAfterExecution = await opcodeTester.rollingBlockDetailComputations();

    logger.debug(`Value before execution: ${valueBeforeExecution}, value after execution: ${valueAfterExecution}`);
    expect(valueBeforeExecution).not.toEqual(valueAfterExecution);

    logger.debug("All opcodes executed successfully");
  });
});
