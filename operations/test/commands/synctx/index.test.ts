import { runCommand } from "@oclif/test";
import { describe, it, expect } from "@jest/globals";

describe("synctx:index", () => {
  it("runs synctx:index cmd", async () => {
    const { stdout } = await runCommand("synctx:index");
    expect(stdout).toContain("hello world");
  });

  it("runs synctx:index --name oclif", async () => {
    const { stdout } = await runCommand("synctx:index --name oclif");
    expect(stdout).toContain("hello oclif");
  });
});
