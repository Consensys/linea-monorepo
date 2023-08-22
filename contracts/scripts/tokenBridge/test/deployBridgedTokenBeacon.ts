import { ethers, upgrades, network } from "hardhat";
import { storeAddress } from "../../../utils/storeAddress";

export async function deployBridgedTokenBeacon(verbose = false) {
  const BridgedToken = await ethers.getContractFactory("BridgedToken");

  const l1TokenBeacon = await upgrades.deployBeacon(BridgedToken);
  await l1TokenBeacon.deployed();
  if (verbose) {
    console.log("L1TokenBeacon deployed, at address:", l1TokenBeacon.address);
  }

  const l2TokenBeacon = await upgrades.deployBeacon(BridgedToken);
  await l2TokenBeacon.deployed();
  if (verbose) {
    console.log("L2TokenBeacon deployed, at address:", l2TokenBeacon.address);
  }

  // @TODO:
  // - Verify contracts on Etherscan

  storeAddress("l1TokenBeacon", l1TokenBeacon.address, network.name);
  storeAddress("l2TokenBeacon", l2TokenBeacon.address, network.name);

  return { l1TokenBeacon, l2TokenBeacon };
}
