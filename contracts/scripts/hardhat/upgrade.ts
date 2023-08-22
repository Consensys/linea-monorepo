/*
Note:
    Deploying an upgradeable contract should generate a hidden directory `.openzeppelin`.
    But because we're deploying from a docker container (or in a cluster) this directory won't be here.

    So you'll first need to run the forceImport script that's in the same directory.

Usage:
    export CONTRACT_NAME=MyContract
    export CONTRACT_ADDRESS=0x1234567890123456789012345678901234567890

    # if you need to specify a different private key then the one in the files
    export PRIVATE_KEY=

    npx hardhat --network <network> run scripts/hardhat/upgrade.js
*/

import { ethers, upgrades } from "hardhat";
import { requireEnv as env } from "./utils";

const CONTRACT_NAME = env("CONTRACT_NAME");
const CONTRACT_ADDRESS = env("CONTRACT_ADDRESS");

async function main() {
  if (!CONTRACT_NAME || !CONTRACT_ADDRESS) {
    throw new Error(`CONTRACT_ADDRESS and CONTRACT_NAME env variables are undefined.`);
  }
  console.log(`Upgrading contract at ${CONTRACT_ADDRESS}`);

  const contract = await ethers.getContractFactory(CONTRACT_NAME);

  console.log("Upgrading...");
  try {
    await upgrades.validateUpgrade(CONTRACT_ADDRESS, contract, {
      kind: "transparent",
    });

    await upgrades.upgradeProxy(CONTRACT_ADDRESS, contract, {
      kind: "transparent",
    });

    console.log(`Upgraded contract at ${CONTRACT_ADDRESS} with a new version of ${CONTRACT_NAME}`);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (error: any) {
    console.error(
      "Failed to upgrade the proxy contract, check the error below. You might have to run the forceImport script first on a pre-upgrade version of the smart contract source to regenerate .openzeppelin files\n",
      error.message,
    );
    throw error;
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
