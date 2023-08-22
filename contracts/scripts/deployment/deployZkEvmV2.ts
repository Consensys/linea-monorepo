import { ethers, run, upgrades } from "hardhat";
import { delay } from "../../utils/storeAddress";
import { requireEnv } from "../hardhat/utils";

/*
    *******************************************************************************************

    *******************************************************************************************
    *******************************************************************************************
    npx hardhat run --network zkevm_dev scripts/deployment/deployZkEvmV2.ts
    *******************************************************************************************
*/

async function main() {
  const proxyAddress = requireEnv("ZKEVMV2_ADDRESS");

  const factory = await ethers.getContractFactory("ZkEvmV2");

  console.log("Deploying V2 Contract...");
  const contractAddress = await upgrades.deployImplementation(factory, {
    kind: "transparent",
  });

  console.log(`Contract deployed at ${contractAddress}`);

  const upgradeCallUsingSecurityCouncil = ethers.utils.hexConcat([
    "0x99a88ec4",
    ethers.utils.defaultAbiCoder.encode(["address", "address"], [proxyAddress, contractAddress]),
  ]);

  console.log(
    "Encoded Tx Upgrade from Security Council:",
    "\n",
    upgradeCallUsingSecurityCouncil
  );

  await delay(120_000);

  await run("verify", {
    address: contractAddress,
  });
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
