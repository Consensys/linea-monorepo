import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { deployFromFactory } from "../scripts/hardhat/utils";
import { getDeploymentSigner, withDeploymentUiSession } from "../scripts/hardhat/deployment-ui";
import { LogContractDeployment, getRequiredEnvVar, tryVerifyContractWithConstructorArgs } from "../common/helpers";

const func: DeployFunction = withDeploymentUiSession(
  "12_deploy_CallForwardingProxy.ts",
  async function (hre: HardhatRuntimeEnvironment) {
    const contractName = "CallForwardingProxy";
    const signer = await getDeploymentSigner(hre);

    // This should be the LineaRollup
    const targetAddress = getRequiredEnvVar("LINEA_ROLLUP_ADDRESS");

    const contract = await deployFromFactory(contractName, signer, targetAddress);

    await LogContractDeployment(contractName, contract);
    const contractAddress = await contract.getAddress();

    const args = [targetAddress];

    await tryVerifyContractWithConstructorArgs(
      contractAddress,
      "contracts/lib/CallForwardingProxy.sol:CallForwardingProxy",
      args,
    );
  },
);
export default func;
func.tags = ["CallForwardingProxy"];
