import { upgrades as createUpgrades } from "@openzeppelin/hardhat-upgrades";
import fs from "fs";
import hre, { network as hardhatNetwork } from "hardhat";
import path from "path";
import { fileURLToPath } from "url";

import { requireEnv } from "../hardhat/utils";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;
const upgrades = await createUpgrades(hre, hardhatConnection);

const currentDir = path.dirname(fileURLToPath(import.meta.url));
const OPENZEPPELIN_DIRECTORY = path.join(currentDir, "..", "..", ".openzeppelin");

async function main() {
  const newContractName = requireEnv("NEW_CONTRACT_NAME");
  const oldContractName = requireEnv("OLD_CONTRACT_NAME");
  const proxyAddress = requireEnv("PROXY_ADDRESS");
  const initialShnarfs = process.env.INITIAL_SHNARFS;
  const initialShnarfsBlockNumbers = process.env.INITIAL_SHNARFS_BLOCK_NUMBERS;

  if (fs.existsSync(OPENZEPPELIN_DIRECTORY)) {
    fs.rmSync(OPENZEPPELIN_DIRECTORY, { recursive: true, force: true });
  }

  const initialShnarfsParam = initialShnarfs ? initialShnarfs.split(",") : [];
  const initialShnarfsBlockNumbersParam = initialShnarfsBlockNumbers ? initialShnarfsBlockNumbers.split(",") : [];

  console.log(`Upgrading contract at ${proxyAddress} from ${oldContractName} to ${newContractName}`);

  const oldContract = await ethers.getContractFactory(oldContractName);
  const newContract = await ethers.getContractFactory(newContractName);

  console.log("Importing contract");
  await upgrades.forceImport(proxyAddress, oldContract, {
    kind: "transparent",
  });

  try {
    await upgrades.upgradeProxy(proxyAddress, newContract, {
      call: {
        fn: "initializeParentShnarfsAndFinalizedState",
        args: [initialShnarfsParam, initialShnarfsBlockNumbersParam],
      },
      kind: "transparent",
      // NOTE: THIS IS ONLY FOR TESTING.
      // TODO: this is because of the transient storage layout change.
      //  Openzeppelin upgradeProxy validation does not allow variable removal even if the storage layout didn't change.
      unsafeSkipStorageCheck: true,
    });

    console.log(`Upgraded contract at ${proxyAddress} from ${oldContractName} to ${newContractName}`);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (error: any) {
    console.error("Failed to upgrade the proxy contract", error.message);
    throw error;
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
