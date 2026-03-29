import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { deployFromFactory } from "../scripts/hardhat/utils";
import { getDeploymentSigner, withDeploymentUiSession } from "../scripts/hardhat/deployment-ui";
import { getRequiredEnvVar, LogContractDeployment, tryVerifyContractWithConstructorArgs } from "../common/helpers";

const func: DeployFunction = withDeploymentUiSession(
  "22_deploy_YieldManagerUpgrade.ts",
  async function (hre: HardhatRuntimeEnvironment) {
    const lineaRollupAddress = getRequiredEnvVar("LINEA_ROLLUP_ADDRESS");

    console.log("Deploying Contract...");
    const signer = await getDeploymentSigner(hre);
    const contractName = "YieldManager";
    const contract = await deployFromFactory(contractName, signer, lineaRollupAddress);
    const yieldManagerAddress = await contract.getAddress();
    await LogContractDeployment(contractName, contract);
    await tryVerifyContractWithConstructorArgs(yieldManagerAddress, "src/yield/YieldManager.sol:YieldManager", [
      lineaRollupAddress,
    ]);
  },
);

export default func;
func.tags = ["YieldManagerImplementation"];
