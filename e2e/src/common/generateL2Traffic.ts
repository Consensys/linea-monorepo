import * as dotenv from "dotenv";
import { config } from "../config/tests-config/setup";
import { etherToWei, sendTransactionsToGenerateTrafficWithInterval } from "./utils";
import { createTestLogger } from "../config/logger";

dotenv.config();

const logger = createTestLogger();

async function main() {
  logger.info("Generating L2 traffic...");

  const pollingAccount = await config.getL2AccountManager().generateAccount(etherToWei("200"));
  const walletClient = config.L2.client.walletClient({ account: pollingAccount });
  const publicClient = config.L2.client.publicClient();

  const stopPolling = await sendTransactionsToGenerateTrafficWithInterval(walletClient, publicClient, {
    pollingInterval: 2_500,
  });

  process.on("SIGINT", () => {
    logger.info("Caught interrupt signal.");
    cleanup();
  });

  process.on("SIGTERM", () => {
    logger.info("Caught termination signal.");
    cleanup();
  });

  function cleanup() {
    stopPolling();
    logger.info("Terminated L2 traffic...");
    process.exit(0);
  }
}

main().catch((error) => {
  logger.error("Error generating L2 traffic", error);
  process.exit(1);
});
