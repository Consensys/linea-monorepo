import { describe, expect, it } from "@jest/globals";
import { config } from "./config/tests-config/setup";
import { L2RpcEndpoint } from "./config/tests-config/setup/clients/l2-client";
import { awaitUntil, execDockerCommand, wait } from "./common/utils";
import { LineaSequencerUptimeFeedAbi } from "./generated";

describe("Liveness test suite", () => {
  it.concurrent(
    "Should succeed to send liveness transactions after sequencer restarted",
    async () => {
      const livenessContract = config.l2PublicClient().getLineaSequencerUptimeFeedContract();

      const latestAnswer = await livenessContract.read.latestAnswer();
      logger.debug(`Latest Status is ${latestAnswer == 1n ? "Down" : "Up"}`);

      let lastBlockTimestamp: bigint = 0n;
      let lastBlockNumber: bigint = 0n;
      const l2BesuNodeClient = config.l2PublicClient({ type: L2RpcEndpoint.BesuNode });
      const sequencerClient = config.l2PublicClient({ type: L2RpcEndpoint.Sequencer });

      try {
        await execDockerCommand("stop", "sequencer");
        logger.debug("Sequencer stopped.");

        // sleep for 9 sec (1 sec longer than the liveness-max-block-age)
        await wait(9000);

        const block = await l2BesuNodeClient.getBlock({ blockTag: "latest" });
        lastBlockTimestamp = block.timestamp;
        lastBlockNumber = block.number;
        logger.debug(`lastBlockNumber=${lastBlockNumber} lastBlockTimestamp=${lastBlockTimestamp}`);

        await execDockerCommand("restart", "sequencer");
        logger.debug("Sequencer restarted.");
      } catch (error) {
        logger.error(`Failed to stop and restart sequencer: ${error}`);
        throw error;
      }

      const targetBlockNumber = lastBlockNumber + 1n;
      logger.debug(`targetBlockNumber=${JSON.stringify(targetBlockNumber)}`);

      let i = 0;
      const livenessEvents = await awaitUntil(
        async () => {
          try {
            logger.debug(`Trial ${i++} to get liveness events`);
            // using fetch JSON-RPC call to get logs instead of JsonRpcProvider to aviod flaky issue
            // where logs would fail to be retrieve from time to time
            return await (i % 2 == 0 ? l2BesuNodeClient : sequencerClient).getLogs({
              address: livenessContract.address,
              fromBlock: targetBlockNumber,
              toBlock: targetBlockNumber,
              events: LineaSequencerUptimeFeedAbi.filter(
                (value) => value.type === "event" && value.name === "AnswerUpdated",
              ),
            });
          } catch (e) {
            return null;
          }
        },
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (ethLogs) => ethLogs != null && ethLogs.length >= 2,
        1000,
        150000,
      );

      logger.debug(`livenessEvents=${JSON.stringify(livenessEvents)}`);
      expect(livenessEvents?.length).toBeGreaterThanOrEqual(2);

      // The first two transactions of the target block should be the transactions
      // with "to" as the liveness contract address
      const targetBlock = await (i % 2 == 0 ? l2BesuNodeClient : sequencerClient).getBlock({
        blockNumber: targetBlockNumber,
        includeTransactions: true,
      });
      logger.debug(`targetBlock=${JSON.stringify(targetBlock)}`);
      expect(targetBlock?.transactions.length).toBeGreaterThanOrEqual(2);

      const downtimeTransaction = targetBlock?.transactions.at(0);
      const uptimeTransaction = targetBlock?.transactions.at(1);
      const downtimeEvent = livenessEvents?.find((event) => event.transactionHash === downtimeTransaction?.hash);
      const uptimeEvent = livenessEvents?.find((event) => event.transactionHash === uptimeTransaction?.hash);

      // check the first AnswerUpdated event is for downtime
      expect(downtimeEvent?.transactionIndex).toEqual("0x0");
      expect(parseInt(downtimeEvent?.topics[1] ?? "", 16)).toEqual(1); // topics[1] was the given status to update, should be 1 for downtime
      expect(parseInt(downtimeEvent?.data ?? "", 16)).toEqual(lastBlockTimestamp); // data should contain the timestamp of the last block before restart as downtime

      // check the second AnswerUpdated event is for uptime
      expect(uptimeEvent?.transactionIndex).toEqual("0x1");
      expect(parseInt(uptimeEvent?.topics[1] ?? "", 16)).toEqual(0); // topics[1] was the given status to update, should be 0 for uptime
      expect(parseInt(uptimeEvent?.data ?? "", 16)).toBeGreaterThan(lastBlockTimestamp ?? 0); // data should contain a timestamp later than the last block before restart as uptime
    },
    150000,
  );
});
