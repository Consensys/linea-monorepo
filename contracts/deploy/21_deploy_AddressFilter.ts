import { ethers } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { LogContractDeployment, getRequiredEnvVar, tryVerifyContractWithConstructorArgs } from "../common/helpers";

const func: DeployFunction = async function () {
  const contractName = "AddressFilter";

  const lineaRollupSecurityCouncil = getRequiredEnvVar("LINEA_ROLLUP_SECURITY_COUNCIL");
  const filteredAddresses = getRequiredEnvVar("ADDRESS_FILTER_FILTERED_ADDRESSES").split(",");

  const factory = await ethers.getContractFactory(contractName);
  const contract = await factory.deploy(lineaRollupSecurityCouncil, filteredAddresses);

  await LogContractDeployment(contractName, contract);
  const contractAddress = await contract.getAddress();

  const args = [lineaRollupSecurityCouncil, filteredAddresses];
  await tryVerifyContractWithConstructorArgs(
    contractAddress,
    "src/rollup/forcedTransactions/AddressFilter.sol:AddressFilter",
    args,
  );
};

export default func;
func.tags = ["AddressFilter"];
