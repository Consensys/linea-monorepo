import { task } from "hardhat/config";

import { getTaskCliOrEnvValue } from "../../../common/helpers/environmentHelper";

/*
    *******************************************************************************************
    1. Set the CONTRACT_TYPE of the proxy - e.g. LineaRollup
    2. Set the PROXY_ADDRESS for the contract 
    *******************************************************************************************

    *******************************************************************************************
    DEPLOYER_PRIVATE_KEY=<key> \
    INFURA_API_KEY=<key> \
    pnpm exec hardhat getCurrentFinalizedBlockNumber \
    --contract-type <string> \
    --proxy-address <address> \
    --network sepolia
    *******************************************************************************************
*/

export default task("getCurrentFinalizedBlockNumber", "Gets the finalized block number")
  .addOption({ name: "contractType", defaultValue: "" })
  .addOption({ name: "proxyAddress", defaultValue: "" })
  .setInlineAction(async (taskArgs, hre) => {
    const { ethers } = await hre.network.getOrCreate();

    const contractType = getTaskCliOrEnvValue(taskArgs, "contractType", "CONTRACT_TYPE");
    const proxyAddress = getTaskCliOrEnvValue(taskArgs, "proxyAddress", "PROXY_ADDRESS");

    if (!contractType || !proxyAddress) {
      throw new Error(`PROXY_ADDRESS and CONTRACT_NAME env variables are undefined.`);
    }

    const LineaRollupContract = await ethers.getContractAt(contractType, proxyAddress);
    const blockNum = await LineaRollupContract.currentL2BlockNumber();

    console.log("Current finalized L2 block number", blockNum);
  })
  .build();
