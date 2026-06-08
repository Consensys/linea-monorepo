import { network as hardhatNetwork } from "hardhat";

import {
  LogContractDeployment,
  getRequiredEnvVar,
  requireAddressFromRegistryOrEnv,
  tryVerifyContractWithConstructorArgs,
} from "../common/helpers";
import { deployScript } from "../rocketh/deploy";
import { getUiSigner, withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";
import { deployFromFactory } from "../scripts/hardhat/utils";
import { get1559Fees } from "../scripts/utils";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;
const networkName = hardhatConnection.networkName === "default" ? "hardhat" : hardhatConnection.networkName;

const func = withSignerUiSession("02_deploy_Timelock.ts", async function () {
  const contractName = "TimeLock";
  const signer = await getUiSigner();

  // This should be the safe
  const timeLockProposers = getRequiredEnvVar("TIMELOCK_PROPOSERS");

  // This should be the safe
  const timelockExecutors = getRequiredEnvVar("TIMELOCK_EXECUTORS");

  // This should be the safe
  const adminAddress = requireAddressFromRegistryOrEnv(networkName, "TIMELOCK_ADMIN_ADDRESS", "TIMELOCK_ADMIN_ADDRESS");

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
});
export default deployScript(func, { tags: ["Timelock"] });
