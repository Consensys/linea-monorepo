import { createTestLogger } from "../logger/logger";

const logger = createTestLogger();

export default async (): Promise<void> => {
  try {
    if (typeof global.stopL2TrafficGeneration === "function") {
      global.stopL2TrafficGeneration();
      logger.debug("Stopped L2 traffic generation");
    } else {
      logger.warn("stopL2TrafficGeneration function not found in global scope");
    }
  } catch (error) {
    logger.error(`Error stopping L2 traffic generation: ${error}`);
    // Don't throw - teardown failures shouldn't mask test failures
  }
};
