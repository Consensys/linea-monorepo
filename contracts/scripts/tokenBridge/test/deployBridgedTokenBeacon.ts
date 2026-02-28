import { ethers } from "../../../test/hardhat/common/connection.js";

export async function deployBridgedTokenBeacon(verbose = false) {
  const BridgedToken = await ethers.getContractFactory("BridgedToken");
  const bridgedTokenImpl = await BridgedToken.deploy();
  await bridgedTokenImpl.waitForDeployment();

  const UpgradeableBeaconFactory = await ethers.getContractFactory("UpgradeableBeacon");

  const l1TokenBeacon = await UpgradeableBeaconFactory.deploy(await bridgedTokenImpl.getAddress());
  await l1TokenBeacon.waitForDeployment();
  if (verbose) {
    console.log("L1TokenBeacon deployed, at address:", await l1TokenBeacon.getAddress());
  }

  const bridgedTokenImpl2 = await BridgedToken.deploy();
  await bridgedTokenImpl2.waitForDeployment();

  const l2TokenBeacon = await UpgradeableBeaconFactory.deploy(await bridgedTokenImpl2.getAddress());
  await l2TokenBeacon.waitForDeployment();
  if (verbose) {
    console.log("L2TokenBeacon deployed, at address:", await l2TokenBeacon.getAddress());
  }

  return { l1TokenBeacon, l2TokenBeacon };
}
