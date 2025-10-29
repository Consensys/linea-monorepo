import { DeployFunction } from "hardhat-deploy/types";
import { deployUpgradableFromFactory } from "../scripts/hardhat/utils";
import { tryVerifyContract, LogContractDeployment } from "../common/helpers";
import { EMPTY_INITIALIZE_SIGNATURE } from "../common/constants";

const func: DeployFunction = async function () {
  const contractName = "RollupRevenueVault";

  const contract = await deployUpgradableFromFactory(contractName, [], {
    initializer: EMPTY_INITIALIZE_SIGNATURE,
    unsafeAllow: ["constructor"],
  });

  await LogContractDeployment(contractName, contract);
  const contractAddress = await contract.getAddress();

  await tryVerifyContract(contractAddress, "src/operational/RollupRevenueVault.sol:RollupRevenueVault");
};

export default func;
func.tags = ["RollupRevenueVault"];
