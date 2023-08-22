import { ethers, upgrades } from "hardhat";
import { requireEnv } from "../hardhat/utils";
import { ZkEvmV2Init__factory } from "../../typechain-types";

/*
    *******************************************************************************************

    *******************************************************************************************
    *******************************************************************************************
    npx hardhat run --network zkevm_dev scripts/deployment/deployImplementation.ts
    *******************************************************************************************
*/

async function main() {
  const proxyAddress = requireEnv("ZKEVMV2_ADDRESS");
  const initialL2BlockNumber = "3";
  const initialStateRootHash = "0x3450000000000000000000000000000000000000000000000000000000000000";

  const factory = await ethers.getContractFactory("ZkEvmV2Init");

  console.log("Deploying V2 Contract...");
  const v2contract = await upgrades.deployImplementation(factory, {
    kind: "transparent",
  });

  console.log(`Contract deployed at ${v2contract}`);

  const upgradeCallWithReinitializationUsingSecurityCouncil = ethers.utils.hexConcat([
    "0x9623609d",
    ethers.utils.defaultAbiCoder.encode(
      ["address", "address", "bytes"],
      [
        proxyAddress,
        v2contract,
        ZkEvmV2Init__factory.createInterface().encodeFunctionData("initializeV2", [
          initialL2BlockNumber,
          initialStateRootHash,
        ]),
      ],
    ),
  ]);

  console.log(
    "Encoded Tx Upgrade with Reinitialization from Security Council:",
    "\n",
    upgradeCallWithReinitializationUsingSecurityCouncil,
  );
  console.log("\n");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
