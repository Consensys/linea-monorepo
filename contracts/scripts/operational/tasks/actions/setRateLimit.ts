import type { NewTaskActionFunction } from "hardhat/types/tasks";
import { getTaskCliOrEnvValue } from "../../../../common/helpers/environmentHelper.js";

interface TaskArgs {
  messageServiceAddress?: string;
  messageServiceType?: string;
  withdrawLimit?: string;
}

const action: NewTaskActionFunction<TaskArgs> = async (taskArgs, hre) => {
  const connection = await hre.network.connect();
  const { ethers } = connection;

  const messageServiceContractType = getTaskCliOrEnvValue(taskArgs, "messageServiceType", "MESSAGE_SERVICE_TYPE");
  const messageServiceAddress = getTaskCliOrEnvValue(taskArgs, "messageServiceAddress", "MESSAGE_SERVICE_ADDRESS");

  if (messageServiceContractType === undefined) {
    throw new Error(
      "Please specify a Message Service name e.g: --message-service-type LineaRollup or MESSAGE_SERVICE_TYPE=LineaRollup",
    );
  }

  if (messageServiceAddress === undefined) {
    throw new Error(
      "Please specify a message service address e.g: --message-service-address 0x... or MESSAGE_SERVICE_ADDRESS=0x...",
    );
  }

  const withdrawLimitInWei = getTaskCliOrEnvValue(taskArgs, "withdrawLimit", "WITHDRAW_LIMIT_IN_WEI");

  if (withdrawLimitInWei === undefined) {
    throw new Error(
      "Please specify a withdraw limit e.g: --withdraw-limit 1000000000000000000 or WITHDRAW_LIMIT_IN_WEI=1000000000000000000",
    );
  }

  const messageService = await ethers.getContractAt(messageServiceContractType, messageServiceAddress);

  const limitInWei = await messageService.limitInWei();
  console.log(
    `Starting with rate limit in wei of ${limitInWei} at ${messageServiceAddress} of type ${messageServiceContractType}`,
  );

  const updateTx = await messageService.resetRateLimitAmount(withdrawLimitInWei);
  await updateTx.wait();

  console.log(`Changed rate limit in wei to ${withdrawLimitInWei} at ${messageServiceAddress}`);

  const newLimitInWei = await messageService.limitInWei();
  console.log(`Validated rate limit in wei of ${newLimitInWei} at ${messageServiceAddress}`);
};

export default action;
