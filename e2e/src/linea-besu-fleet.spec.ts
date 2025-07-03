import { ethers, JsonRpcProvider } from "ethers";
import { describe, expect, it } from "@jest/globals";
import { config } from "./config/tests-config";
import { LineaEstimateGasClient } from "./common/utils";

const l2AccountManager = config.getL2AccountManager();

describe("Linea besu fleet test suite", () => {
  const lineaEstimateGasLeaderClient = new LineaEstimateGasClient(config.getL2BesuNodeEndpoint()!);
  const lineaEstimateGasFollowerClient = new LineaEstimateGasClient(config.getL2BesuFollowerNodeEndpoint()!);
  const leaderL2Provider = new JsonRpcProvider(config.getL2BesuNodeEndpoint()!.toString());
  const followerL2Provider = new JsonRpcProvider(config.getL2BesuFollowerNodeEndpoint()!.toString());

  it.concurrent("linea_estimateGas responses from leader and follower should match", async () => {
    const account = await l2AccountManager.generateAccount();
    const dummyContract = config.getL2DummyContract(account);
    const randomBytes = ethers.randomBytes(1000);

    const gasFeesFromLeader = await lineaEstimateGasLeaderClient.lineaEstimateGas(
      account.address,
      await dummyContract.getAddress(),
      dummyContract.interface.encodeFunctionData("setPayload", [randomBytes]),
    );
    logger.debug(
      `Fetched fee data from leader. maxPriorityFeePerGas=${gasFeesFromLeader.maxPriorityFeePerGas} maxFeePerGas=${gasFeesFromLeader.maxFeePerGas}`,
    );

    const gasFeesFromFollower = await lineaEstimateGasFollowerClient.lineaEstimateGas(
      account.address,
      await dummyContract.getAddress(),
      dummyContract.interface.encodeFunctionData("setPayload", [randomBytes]),
    );
    logger.debug(
      `Fetched fee data from follower. maxPriorityFeePerGas=${gasFeesFromFollower.maxPriorityFeePerGas} maxFeePerGas=${gasFeesFromFollower.maxFeePerGas}`,
    );

    expect(gasFeesFromLeader.maxPriorityFeePerGas).toEqual(gasFeesFromFollower.maxPriorityFeePerGas);
    expect(gasFeesFromLeader.maxFeePerGas).toEqual(gasFeesFromFollower.maxFeePerGas);
  });

  it.concurrent("eth_estimateGas responses from leader and follower should match", async () => {
    const account = await l2AccountManager.generateAccount();
    const dummyContract = config.getL2DummyContract(account);
    const randomBytes = ethers.randomBytes(1000);

    const estimatedGasFromLeader = await leaderL2Provider.estimateGas({
      from: account.address,
      to: await dummyContract.getAddress(),
      data: dummyContract.interface.encodeFunctionData("setPayload", [randomBytes]),
    });
    logger.debug(`Fetched fee data from leader. estimatedGasFromLeader=${estimatedGasFromLeader}`);

    const estimatedGasFromFollower = await followerL2Provider.estimateGas({
      from: account.address,
      to: await dummyContract.getAddress(),
      data: dummyContract.interface.encodeFunctionData("setPayload", [randomBytes]),
    });
    logger.debug(`Fetched fee data from follower. estimatedGasFromFollower=${estimatedGasFromFollower}`);

    expect(estimatedGasFromLeader).toEqual(estimatedGasFromFollower);
  });

  it.concurrent("Finalized block numbers and hashes from leader and follower should match", async () => {
    const getBlockFromLeader = await leaderL2Provider.getBlock("finalized", false);
    const getBlockFromFollower = await followerL2Provider.getBlock("finalized", false);

    logger.debug(
      `Fetched finalized block and hash from leader. finalizedBlockNumber=${getBlockFromLeader!.number} hash=${getBlockFromLeader!.hash}`,
    );
    logger.debug(
      `Fetched finalized block and hash from follower. finalizedBlockNumber=${getBlockFromFollower!.number} hash=${getBlockFromFollower!.hash}`,
    );

    expect(getBlockFromLeader!.number).toEqual(getBlockFromFollower!.number);
    expect(getBlockFromLeader!.hash).toEqual(getBlockFromFollower!.hash);
  });
});
