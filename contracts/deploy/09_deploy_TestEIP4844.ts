import { ethers } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { deployFromFactory } from "../scripts/hardhat/utils";
import { get1559Fees } from "../scripts/utils";

const func: DeployFunction = async function () {
  const contractName = "TestEIP4844";

  const provider = ethers.provider;

  const contract = await deployFromFactory(contractName, provider, await get1559Fees(provider));
  const contractAddress = await contract.getAddress();

  const txReceipt = await contract.deploymentTransaction()?.wait();
  if (!txReceipt) {
    throw "Deployment transaction not found.";
  }

  const chainId = (await ethers.provider!.getNetwork()).chainId;
  console.log(
    `contract=${contractName} deployed: address=${contractAddress} blockNumber=${txReceipt.blockNumber} chainId=${chainId}`,
  );
};
export default func;
func.tags = ["TestEIP4844"];
