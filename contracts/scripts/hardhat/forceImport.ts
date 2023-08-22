/*
Note:
    Make sure to run this script BEFORE making any changes to the smart contract.
    The script will generate a hidden directory `.openzeppelin` that will be used by the upgrade script.
    But this directory must contain data generated from the currently deployed version of the smart contract.

Usage:
    export CONTRACT_NAME=MyContract
    export CONTRACT_ADDRESS=0x1234567890123456789012345678901234567890

    npx hardhat --network <network> run scripts/hardhat/forceImport.js
*/

import { ethers, upgrades } from "hardhat";
import { requireEnv as env } from "./utils";

const CONTRACT_NAME = env("CONTRACT_NAME");
const CONTRACT_ADDRESS = env("CONTRACT_ADDRESS");

async function main() {
  if (!CONTRACT_NAME || !CONTRACT_ADDRESS) {
    throw new Error(`CONTRACT_ADDRESS and CONTRACT_NAME env variables are undefined.`);
  }
  console.log(`Importing contract at ${CONTRACT_ADDRESS}`);

  const contract = await ethers.getContractFactory(CONTRACT_NAME);

  console.log("Importing contract");
  await upgrades.forceImport(CONTRACT_ADDRESS, contract, {
    kind: "transparent",
  });
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
