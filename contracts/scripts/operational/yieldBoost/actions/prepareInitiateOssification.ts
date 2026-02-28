import type { NewTaskActionFunction } from "hardhat/types/tasks";
import { getTaskCliOrEnvValue } from "../../../../common/helpers/environmentHelper.js";

interface TaskArgs {
  yieldManager?: string;
  yieldProvider?: string;
}

const action: NewTaskActionFunction<TaskArgs> = async (taskArgs, hre) => {
  const connection = await hre.network.connect();
  const { ethers } = connection;

  const yieldManager = getTaskCliOrEnvValue(taskArgs, "yieldManager", "YIELD_MANAGER_ADDRESS");
  const yieldProvider = getTaskCliOrEnvValue(taskArgs, "yieldProvider", "YIELD_PROVIDER_ADDRESS");

  if (!yieldManager) {
    throw new Error("Please specify yieldManager / YIELD_MANAGER_ADDRESS");
  }
  if (!yieldProvider) {
    throw new Error("Please specify yieldProvider / YIELD_PROVIDER_ADDRESS");
  }

  const yieldProviderContract = await ethers.getContractAt("LidoStVaultYieldProvider", yieldProvider);
  const isOssified = await yieldProviderContract.isOssified();

  console.log("YieldProvider:", yieldProvider);
  console.log("Is Ossified:", isOssified);
  console.log("\nTo initiate ossification, call initiateOssification() on the YieldProvider");
};

export default action;
