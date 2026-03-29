import { ethers } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { deployFromFactory } from "../scripts/hardhat/utils";
import { getDeploymentSigner, withDeploymentUiSession } from "../scripts/hardhat/deployment-ui";
import { get1559Fees } from "../scripts/utils";
import { LogContractDeployment, getRequiredEnvVar, tryVerifyContractWithConstructorArgs } from "../common/helpers";

const func: DeployFunction = withDeploymentUiSession(
  "02_deploy_Timelock.ts",
  async function (hre: HardhatRuntimeEnvironment) {
    const contractName = "TimeLock";
    const signer = await getDeploymentSigner(hre);

    // This should be the safe
    const timeLockProposers = getRequiredEnvVar("TIMELOCK_PROPOSERS");

    // This should be the safe
    const timelockExecutors = getRequiredEnvVar("TIMELOCK_EXECUTORS");

    // This should be the safe
    const adminAddress = getRequiredEnvVar("TIMELOCK_ADMIN_ADDRESS");

    const minDelay = process.env.MIN_DELAY || 0;

    const contract = await deployFromFactory(
      contractName,
      signer,
      minDelay,
      timeLockProposers?.split(","),
      timelockExecutors?.split(","),
      adminAddress,
      await get1559Fees(ethers.provider),
    );

    await LogContractDeployment(contractName, contract);
    const contractAddress = await contract.getAddress();

    const args = [minDelay, timeLockProposers?.split(","), timelockExecutors?.split(","), adminAddress];

    await tryVerifyContractWithConstructorArgs(contractAddress, "src/governance/TimeLock.sol:TimeLock", args);
  },
);
export default func;
func.tags = ["Timelock"];
