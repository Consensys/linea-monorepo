import { runCommand } from "@oclif/test";
import { describe, it, expect } from "@jest/globals";

describe("eth-transfer:index", () => {
  it("runs eth-transfer:index cmd", async () => {
    const { stdout } = await runCommand("eth-transfer:index");
    expect(stdout).toContain("hello world");
  });

  it("runs eth-transfer:index --name oclif", async () => {
    const { stdout } = await runCommand("eth-transfer:index --name oclif");
    expect(stdout).toContain("hello oclif");
  });
});
