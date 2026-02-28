import type { NewTaskActionFunction } from "hardhat/types/tasks";
import { getTaskCliOrEnvValue } from "../../../../common/helpers/environmentHelper.js";

interface TaskArgs {
  tokenBridgeAddress?: string;
  newMessageServiceAddress?: string;
}

const action: NewTaskActionFunction<TaskArgs> = async (taskArgs, hre) => {
  const connection = await hre.network.connect();
  const { ethers } = connection;

  const tokenBridgeAddress = getTaskCliOrEnvValue(taskArgs, "tokenBridgeAddress", "TOKEN_BRIDGE_ADDRESS");
  const newMessageServiceAddress = getTaskCliOrEnvValue(
    taskArgs,
    "newMessageServiceAddress",
    "NEW_MESSAGE_SERVICE_ADDRESS",
  );

  if (tokenBridgeAddress === undefined) {
    throw new Error(
      "Please specify a token bridge address e.g: --token-bridge-address 0x... or TOKEN_BRIDGE_ADDRESS=0x...",
    );
  }

  if (newMessageServiceAddress === undefined) {
    throw new Error(
      "Please specify a new message service address e.g: --new-message-service-address 0x... or NEW_MESSAGE_SERVICE_ADDRESS=0x...",
    );
  }

  const tokenBridge = await ethers.getContractAt("TokenBridge", tokenBridgeAddress);

  console.log(`Setting message service to ${newMessageServiceAddress} on TokenBridge at ${tokenBridgeAddress}`);
  const tx = await tokenBridge.setMessageService(newMessageServiceAddress);
  console.log("Waiting for transaction to process");
  await tx.wait();

  console.log("Done");
};

export default action;
