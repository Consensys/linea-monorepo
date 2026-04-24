import { ethers } from "hardhat";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction } from "hardhat-deploy/types";

import { getRequiredEnvVar, LogContractDeployment, tryVerifyContractWithConstructorArgs } from "../common/helpers";
import { getUiSigner, withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";
import { deployFromFactory } from "../scripts/hardhat/utils";

const func: DeployFunction = withSignerUiSession(
  "22_deploy_YieldManagerUpgrade.ts",
  async function (hre: HardhatRuntimeEnvironment) {
    const lineaRollupAddress = getRequiredEnvVar("LINEA_ROLLUP_ADDRESS");
    const yieldManagerProxyAddress = getRequiredEnvVar("YIELD_MANAGER_ADDRESS");

    console.log("Deploying Contract...");
    const signer = await getUiSigner(hre);
    const contractName = "YieldManager";
    const contract = await deployFromFactory(contractName, signer, lineaRollupAddress);
    const yieldManagerAddress = await contract.getAddress();
    await LogContractDeployment(contractName, contract);
    await tryVerifyContractWithConstructorArgs(yieldManagerAddress, "src/yield/YieldManager.sol:YieldManager", [
      lineaRollupAddress,
    ]);

    // Encoding for the upgrade call to be executed through the Safe.
    // THIS IS JUST A SAMPLE AND WILL BE ADJUSTED WHEN NEEDED FOR GENERATING THE CALLDATA FOR THE UPGRADE CALL
    // https://www.4byte.directory/signatures/?bytes4_signature=0x99a88ec4
    const upgradeCallUsingSecurityCouncil = ethers.concat([
      "0x99a88ec4",
      ethers.AbiCoder.defaultAbiCoder().encode(["address", "address"], [yieldManagerProxyAddress, yieldManagerAddress]),
    ]);

    console.log("Encoded Tx Upgrade from Security Council:", "\n", upgradeCallUsingSecurityCouncil);
    console.log("\n");
  },
);

export default func;
func.tags = ["YieldManagerImplementation"];
