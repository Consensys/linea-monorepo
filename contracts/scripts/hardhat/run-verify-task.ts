const VERIFY_BIGINT_SENTINEL = "__linea_verify_bigint__";

function parseVerifyArgs(rawArgs: string): Record<string, unknown> {
  return JSON.parse(rawArgs, (_key, value: unknown) => {
    const maybeSerializedBigInt =
      value !== null && typeof value === "object" && !Array.isArray(value) ? (value as Record<string, unknown>) : null;

    if (
      maybeSerializedBigInt !== null &&
      Object.keys(maybeSerializedBigInt).length === 1 &&
      typeof maybeSerializedBigInt[VERIFY_BIGINT_SENTINEL] === "string"
    ) {
      return BigInt(maybeSerializedBigInt[VERIFY_BIGINT_SENTINEL] as string);
    }
    return value;
  }) as Record<string, unknown>;
}

async function main() {
  const task = process.env.HARDHAT_VERIFY_TASK;
  const rawArgs = process.env.HARDHAT_VERIFY_ARGS;

  if (!task) {
    throw new Error("Missing HARDHAT_VERIFY_TASK.");
  }
  if (!rawArgs) {
    throw new Error("Missing HARDHAT_VERIFY_ARGS.");
  }

  const args = parseVerifyArgs(rawArgs);
  const hreModule = await import("hardhat");
  const run =
    hreModule.run ??
    (
      hreModule.default as
        | { run?: ((taskName: string, taskArgs?: Record<string, unknown>) => Promise<unknown>) | undefined }
        | undefined
    )?.run;

  if (!run) {
    throw new Error("Hardhat runtime not available in verification child process.");
  }

  await run(task, args);
}

main()
  .then(() => {
    process.exit(0);
  })
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
