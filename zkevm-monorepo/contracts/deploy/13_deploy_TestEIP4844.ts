import { ethers } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { deployFromFactory } from "../scripts/hardhat/utils";
import { get1559Fees } from "../scripts/utils";

const func: DeployFunction = async function () {
  const contractName = "TestEIP4844";

  const provider = ethers.provider;

  const contract = await deployFromFactory(contractName, provider, await get1559Fees(provider));
  const contractAddress = await contract.getAddress();

  console.log(`${contractName} deployed at ${contractAddress}`);

  const deployTx = contract.deploymentTransaction();
  if (!deployTx) {
    throw "Contract deployment transaction receipt not found.";
  }
};
export default func;
func.tags = ["TestEIP4844"];
