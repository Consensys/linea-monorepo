import { ethers } from "hardhat";
import { requireEnv } from "../hardhat/utils";

/*
    *******************************************************************************************
    1. Set the CONTRACT_TYPE of the proxy - e.g. ZkEvmV2
    2. Set the PROXY_ADDRESS for the contract 
    *******************************************************************************************

    *******************************************************************************************
    npx hardhat run --network zkevm_dev scripts/operational/getCurrentFinalizedBlockNumber.ts
    *******************************************************************************************
*/

async function main() {
  const contractType = requireEnv("CONTRACT_TYPE");
  const proxyAddress = requireEnv("PROXY_ADDRESS");

  if (!contractType || !proxyAddress) {
    throw new Error(`PROXY_ADDRESS and CONTRACT_NAME env variables are undefined.`);
  }

  const zkEvmContract = await ethers.getContractAt(contractType, proxyAddress);
  const blockNum = await zkEvmContract.currentL2BlockNumber();

  console.log(blockNum);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
