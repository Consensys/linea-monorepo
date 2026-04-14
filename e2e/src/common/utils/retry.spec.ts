import { describe, expect, it } from "@jest/globals";

import { isContractRevert } from "./retry";

describe("retry helper", () => {
  it("detects transaction execution contract reverts", () => {
    expect(
      isContractRevert({
        name: "TransactionExecutionError",
        cause: { name: "ContractFunctionRevertedError" },
      }),
    ).toBe(true);
  });

  it("detects contract function execution reverts from viem simulation", () => {
    expect(
      isContractRevert({
        name: "ContractFunctionExecutionError",
        cause: { name: "ContractFunctionRevertedError" },
      }),
    ).toBe(true);
  });

  it("detects nested call execution reverts", () => {
    expect(
      isContractRevert({
        name: "ContractFunctionExecutionError",
        cause: {
          name: "EstimateGasExecutionError",
          cause: { name: "CallExecutionError" },
        },
      }),
    ).toBe(true);
  });

  it("ignores non-revert send errors", () => {
    expect(
      isContractRevert({
        name: "TransactionExecutionError",
        cause: { name: "NonceTooLowError" },
      }),
    ).toBe(false);
  });
});
