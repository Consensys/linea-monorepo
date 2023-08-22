import { ethers } from "hardhat";
import { deployFromFactory, requireEnv } from "../hardhat/utils";
import { get1559Fees } from "../utils";

/*
    *******************************************************************************************
    1. Set the Safe address for TIMELOCK_PROPOSERS
    2. Set the Safe address for TIMELOCK_EXECUTORS
    3. Set the Safe address for the TIMELOCK_ADMIN_ADDRESS / optional see note below
    *******************************************************************************************
    IMPORTANT: The optional admin can aid with initial configuration of roles after deployment
    without being subject to delay, but this role should be subsequently renounced in favor of
    administration through timelocked proposals. Previous versions of this contract would assign
    this admin to the deployer automatically and should be renounced as well.
    *******************************************************************************************
    npx hardhat run --network zkevm_dev scripts/deployment/deployTimelock.ts
    *******************************************************************************************
*/

async function main() {
  const provider = ethers.provider;

  // This should be the safe
  const timeLockProposers = requireEnv("TIMELOCK_PROPOSERS");

  // This should be the safe
  const timelockExecutors = requireEnv("TIMELOCK_EXECUTORS");

  // This should be the safe
  const adminAddress = requireEnv("TIMELOCK_ADMIN_ADDRESS");

  const minDelay = process.env.MIN_DELAY || 0;

  const timelock = await deployFromFactory(
    "TimeLock",
    provider,
    minDelay,
    timeLockProposers?.split(","),
    timelockExecutors?.split(","),
    adminAddress,
    await get1559Fees(provider),
  );

  console.log("TimeLock deployed to:", timelock.address);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
