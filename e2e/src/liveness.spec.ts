import { describe, expect, it } from "@jest/globals";
import { config } from "./config/tests-config";
import { execDockerCommand, getBlockByNumberOrBlockTag, pollForBlockNumber, wait } from "./common/utils";

// should remove skip only when the linea-sequencer plugin supports liveness
describe.skip("Liveness test suite", () => {
  const l2AccountManager = config.getL2AccountManager();

  it.concurrent(
    "Should succeed to send liveness transactions after sequencer restared",
    async () => {
      const account = await l2AccountManager.generateAccount();
      const livenessContract = config.getL2LineaSequencerUptimeFeedContract(account);

      const latestAnswer = await livenessContract.latestAnswer();
      logger.debug(`Latest Status is ${latestAnswer == 1n ? true : false}`);

      let lastBlockTimestamp: number | undefined = 0;
      let lastBlockNumber: number | undefined = 0;

      try {
        await execDockerCommand("stop", "sequencer");
        logger.debug("Sequencer stopped.");

        // sleep for 8 sec
        await wait(8000);

        const block = await config.getL2Provider().getBlock("latest");
        lastBlockTimestamp = block?.timestamp;
        lastBlockNumber = block?.number;
        logger.debug(`lastBlockNumber=${lastBlockNumber} lastBlockTimestamp=${lastBlockTimestamp}`);

        await execDockerCommand("restart", "sequencer");
        logger.debug("Sequencer restarted.");
      } catch (error) {
        logger.error(`Failed to stop and restart sequencer: ${error}`);
        throw error;
      }

      await pollForBlockNumber(config.getL2Provider(), lastBlockNumber! + 1);

      const targetBlock = await getBlockByNumberOrBlockTag(config.getL2BesuNodeEndpoint()!, lastBlockNumber! + 1, true);
      const targetBlockTimestamp: number | undefined = targetBlock?.timestamp;
      logger.debug(`targetBlock=${JSON.stringify(targetBlock)}`);

      // The latest status should be true to indicate sequencer is up
      // and the startedAt and updatedAt should be greater than or equal to the target block timestamp
      const latestRoundData = await livenessContract.latestRoundData();
      expect(latestRoundData.answer).toEqual(1n);
      expect(latestRoundData.startedAt).toBeGreaterThanOrEqual(BigInt(targetBlockTimestamp!));
      expect(latestRoundData.updatedAt).toBeGreaterThanOrEqual(BigInt(targetBlockTimestamp!));

      // The first two transactions of the target block should be the transactions
      // with "to" as the liveness contract address
      expect(targetBlock?.transactions.length).toBeGreaterThanOrEqual(2);
      expect((await targetBlock?.getTransaction(0))?.to).toEqual(
        await config.getL2LineaSequencerUptimeFeedContract().getAddress(),
      );
      expect((await targetBlock?.getTransaction(1))?.to).toEqual(
        await config.getL2LineaSequencerUptimeFeedContract().getAddress(),
      );
    },
    60000,
  );
});
