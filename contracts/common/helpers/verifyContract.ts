import { spawn } from "node:child_process";
import { resolve } from "node:path";
import { clearUiWorkflowStatus, setUiWorkflowStatus } from "../../scripts/hardhat/signer-ui-bridge";
import { delay } from "./general";

const VERIFY_TIMEOUT_MS = 90_000;
const VERIFY_PROPAGATION_DELAY_MS = 30_000;
const VERIFY_CHILD_SCRIPT = resolve(__dirname, "../../scripts/hardhat/run-verify-task.ts");
const VERIFY_BIGINT_SENTINEL = "__linea_verify_bigint__";

function stringifyVerifyArgs(args: Record<string, unknown>): string {
  return JSON.stringify(args, (_key, value: unknown) => {
    if (typeof value === "bigint") {
      return {
        [VERIFY_BIGINT_SENTINEL]: value.toString(),
      };
    }
    return value;
  });
}

async function getCurrentHardhatNetworkName(): Promise<string> {
  const hreModule = await import("hardhat");
  const networkName =
    hreModule.network?.name ?? (hreModule.default as { network?: { name?: string } } | undefined)?.network?.name;

  if (!networkName) {
    throw new Error("Hardhat network name is not available; ensure verification runs under Hardhat.");
  }

  return networkName;
}

async function runVerifyTaskWithTimeout(
  task: string,
  args: Record<string, unknown>,
): Promise<"completed" | "timed_out"> {
  const networkName = await getCurrentHardhatNetworkName();

  return await new Promise<"completed" | "timed_out">((resolveRun) => {
    const child = spawn(
      "pnpm",
      ["exec", "hardhat", "run", "--no-compile", "--network", networkName, VERIFY_CHILD_SCRIPT],
      {
        cwd: resolve(__dirname, "../.."),
        env: {
          ...process.env,
          HARDHAT_VERIFY_TASK: task,
          HARDHAT_VERIFY_ARGS: stringifyVerifyArgs(args),
        },
        stdio: "inherit",
      },
    );

    let settled = false;
    const finish = (result: "completed" | "timed_out") => {
      if (settled) {
        return;
      }
      settled = true;
      clearTimeout(timeout);
      resolveRun(result);
    };

    const timeout = setTimeout(() => {
      try {
        child.kill("SIGTERM");
      } catch {
        /* child may already be gone */
      }
      setTimeout(() => {
        try {
          child.kill("SIGKILL");
        } catch {
          /* child may already be gone */
        }
      }, 1000).unref();
      finish("timed_out");
    }, VERIFY_TIMEOUT_MS);

    child.once("exit", () => {
      finish("completed");
    });
    child.once("error", (error) => {
      console.log(`Error happened during verification: ${error}`);
      finish("completed");
    });
  });
}

async function verifyBestEffort(task: string, args: Record<string, unknown>, contractAddress: string): Promise<void> {
  setUiWorkflowStatus(
    "waiting_for_contract_verification",
    `Waiting for contract verification for ${contractAddress}. Explorer propagation may take around 30 seconds.`,
  );
  console.log("Waiting 30 seconds for contract propagation...");
  await delay(VERIFY_PROPAGATION_DELAY_MS);
  console.log("Etherscan verification ongoing...");

  try {
    const result = await runVerifyTaskWithTimeout(task, args);
    if (result === "timed_out") {
      console.log(
        `Verification timed out after ${VERIFY_TIMEOUT_MS / 1000}s. Continuing deploy; you can verify ${contractAddress} separately later.`,
      );
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.log(`Verification failed for ${contractAddress}: ${message}`);
    console.log(`Continuing deploy; you can verify ${contractAddress} separately later.`);
  } finally {
    clearUiWorkflowStatus();
  }

  console.log("Etherscan verification done.");
}

export async function tryVerifyContract(contractAddress: string, contractForVerification?: string) {
  if (process.env.VERIFY_CONTRACT === "true") {
    const verifyArgs: Record<string, unknown> = {
      address: contractAddress,
    };
    if (contractForVerification) {
      verifyArgs.contract = contractForVerification;
    }
    await verifyBestEffort("verify", verifyArgs, contractAddress);
  }
}

export async function tryVerifyContractWithConstructorArgs(
  contractAddress: string,
  contractForVerification: string,
  args: unknown[],
  libraries?: Record<string, string>,
) {
  if (process.env.VERIFY_CONTRACT === "true") {
    await verifyBestEffort(
      "verify:verify",
      {
        address: contractAddress,
        contract: contractForVerification,
        constructorArguments: args,
        libraries: libraries,
      },
      contractAddress,
    );
  }
}
