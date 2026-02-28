import type { NewTaskActionFunction } from "hardhat/types/tasks";
import { getTaskCliOrEnvValue } from "../../../../common/helpers/environmentHelper.js";

interface TaskArgs {
  contractType?: string;
  proxyAddress?: string;
}

const action: NewTaskActionFunction<TaskArgs> = async (taskArgs, hre) => {
  const connection = await hre.network.connect();
  const { ethers } = connection;

  const contractType = getTaskCliOrEnvValue(taskArgs, "contractType", "CONTRACT_TYPE");
  const proxyAddress = getTaskCliOrEnvValue(taskArgs, "proxyAddress", "PROXY_ADDRESS");

  if (!contractType || !proxyAddress) {
    throw new Error(`PROXY_ADDRESS and CONTRACT_TYPE env variables are undefined.`);
  }

  const LineaRollupContract = await ethers.getContractAt(contractType, proxyAddress);
  const blockNum = await LineaRollupContract.currentL2BlockNumber();

  console.log("Current finalized L2 block number", blockNum);
};

export default action;
