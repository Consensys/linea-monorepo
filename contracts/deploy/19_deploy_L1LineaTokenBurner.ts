import { ethers } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { getDeploymentSigner, withDeploymentUiSession } from "../scripts/hardhat/deployment-ui";
import { getRequiredEnvVar, LogContractDeployment, tryVerifyContractWithConstructorArgs } from "../common/helpers";

const func: DeployFunction = withDeploymentUiSession(
  "19_deploy_L1LineaTokenBurner.ts",
  async function (hre: HardhatRuntimeEnvironment) {
    const contractName = "L1LineaTokenBurner";
    const signer = await getDeploymentSigner(hre);

    const messageService = getRequiredEnvVar("LINEA_ROLLUP_ADDRESS");
    const lineaToken = getRequiredEnvVar("LINEA_TOKEN_BURNER_LINEA_TOKEN");

    const factory = await ethers.getContractFactory(contractName, signer);
    const contract = await factory.connect(signer).deploy(messageService, lineaToken);

    await LogContractDeployment(contractName, contract);
    const contractAddress = await contract.getAddress();

    const args = [messageService, lineaToken];
    await tryVerifyContractWithConstructorArgs(
      contractAddress,
      "src/operational/L1LineaTokenBurner.sol:L1LineaTokenBurner",
      args,
    );
  },
);

export default func;
func.tags = ["L1LineaTokenBurner"];
