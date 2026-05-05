import { runCommand } from "@oclif/test";
import { resolve } from "path";

const operationsRoot = resolve(__dirname, "../..");

export function normalizeWhitespace(input: string) {
  return input.replace(/\s+/g, " ");
}

export async function runOperationsCommand(args: string[]) {
  const result = await runCommand(args, { root: operationsRoot }, { testNodeEnv: "production" });
  process.exitCode = 0;
  return result;
}
