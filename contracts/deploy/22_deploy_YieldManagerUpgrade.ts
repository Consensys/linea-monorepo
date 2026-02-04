import { DeployFunction } from "hardhat-deploy/types";
import { ethers } from "hardhat";
import { deployFromFactory } from "../scripts/hardhat/utils";
import { getRequiredEnvVar, LogContractDeployment, tryVerifyContractWithConstructorArgs } from "../common/helpers";

const func: DeployFunction = async function (hre) {
  const lineaRollupAddress = getRequiredEnvVar("LINEA_ROLLUP_ADDRESS");

  console.log("Deploying Contract...");
  const provider = ethers.provider;
  const contractName = "YieldManager";
  const contract = await deployFromFactory(contractName, provider, lineaRollupAddress);
  const yieldManagerAddress = await contract.getAddress();
  await LogContractDeployment(contractName, contract);
  await tryVerifyContractWithConstructorArgs(hre.run, yieldManagerAddress, "src/yield/YieldManager.sol:YieldManager", [
    lineaRollupAddress,
  ]);
};

export default func;
func.tags = ["YieldManagerImplementation"];
