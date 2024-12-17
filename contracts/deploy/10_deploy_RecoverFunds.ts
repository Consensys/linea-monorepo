import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { deployUpgradableFromFactory } from "../scripts/hardhat/utils";
import { tryVerifyContract, getDeployedContractAddress, tryStoreAddress, getRequiredEnvVar } from "../common/helpers";
import { ethers } from "hardhat";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments } = hre;

  const contractName = "RecoverFunds";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);

  // RecoverFunds DEPLOYED AS UPGRADEABLE PROXY
  const RecoverFunds_securityCouncil = getRequiredEnvVar("RECOVERFUNDS_SECURITY_COUNCIL");
  const RecoverFunds_executorAddress = getRequiredEnvVar("RECOVERFUNDS_EXECUTOR_ADDRESS");

  console.log(`Setting security council ${RecoverFunds_securityCouncil}`);
  console.log(`Setting executor address ${RecoverFunds_executorAddress}`);

  if (existingContractAddress === undefined) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }
  const contract = await deployUpgradableFromFactory(
    "RecoverFunds",
    [RecoverFunds_securityCouncil, RecoverFunds_executorAddress],
    {
      initializer: "initialize(address, address)",
      unsafeAllow: ["constructor"],
    },
  );

  const contractAddress = await contract.getAddress();
  const txReceipt = await contract.deploymentTransaction()?.wait();
  if (!txReceipt) {
    throw "Contract deployment transaction receipt not found.";
  }

  const chainId = (await ethers.provider!.getNetwork()).chainId;
  console.log(
    `contract=${contractName} deployed: address=${contractAddress} blockNumber=${txReceipt.blockNumber} chainId=${chainId}`,
  );

  await tryStoreAddress(hre.network.name, contractName, contractAddress, txReceipt.hash);

  await tryVerifyContract(contractAddress);
};

export default func;
func.tags = ["RecoverFunds"];
