import { createTestLogger } from "../logger";
import { ensureOnceOffPrerequisites } from "./prerequisites";
import { startL2TrafficGeneration } from "./traffic";
import { setStopL2TrafficGeneration } from "./traffic-state";
import { createTestContext } from "../tests-config/setup";

const logger = createTestLogger();

export default async (): Promise<void> => {
  const context = createTestContext();

  await ensureOnceOffPrerequisites(context, logger);

  logger.info("Generating L2 traffic...");
  const stopPolling = await startL2TrafficGeneration(context, { pollingIntervalMs: 2_000 });

  setStopL2TrafficGeneration(stopPolling);
};
