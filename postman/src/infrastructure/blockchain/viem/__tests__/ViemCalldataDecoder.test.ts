import { describe, it, expect } from "@jest/globals";
import { encodeFunctionData, parseAbi } from "viem";

import { ViemCalldataDecoder } from "../ViemCalldataDecoder";

describe("ViemCalldataDecoder", () => {
  const decoder = new ViemCalldataDecoder();

  it("decodes a simple uint256 argument", () => {
    const abi = parseAbi(["function transfer(address to, uint256 amount)"]);
    const calldata = encodeFunctionData({
      abi,
      functionName: "transfer",
      args: ["0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", 1000n],
    });

    const result = decoder.decode("function transfer(address to, uint256 amount)", calldata);

    expect((result["0"] as string).toLowerCase()).toBe("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb");
    expect(result["1"]).toBe(1000n);
  });

  it("decodes a function with no arguments by returning empty record", () => {
    const abi = parseAbi(["function noArgs()"]);
    const calldata = encodeFunctionData({ abi, functionName: "noArgs" });

    const result = decoder.decode("function noArgs()", calldata);
    expect(result).toEqual({});
  });

  it("decodes bytes calldata argument", () => {
    const abi = parseAbi(["function sendMessage(address to, bytes calldata)"]);
    const calldata = encodeFunctionData({
      abi,
      functionName: "sendMessage",
      args: ["0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "0xdeadbeef"],
    });

    const result = decoder.decode("function sendMessage(address to, bytes calldata)", calldata);
    expect((result["0"] as string).toLowerCase()).toBe("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa");
    expect(result["1"]).toBe("0xdeadbeef");
  });

  it("accepts human-readable ABI string matching ethers Interface format", () => {
    // Verify that parseAbi can handle the same format that ethers Interface accepted
    const iface = "function foo(uint256 bar, address baz)";
    const abi = parseAbi([iface]);
    const calldata = encodeFunctionData({
      abi,
      functionName: "foo",
      args: [42n, "0xcccccccccccccccccccccccccccccccccccccccc"],
    });

    const result = decoder.decode(iface, calldata);
    expect(result["0"]).toBe(42n);
    expect((result["1"] as string).toLowerCase()).toBe("0xcccccccccccccccccccccccccccccccccccccccc");
  });
});
