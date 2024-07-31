import { ethers, network, upgrades } from "hardhat";
import { tryStoreAddress } from "../../../utils/storeAddress";
import { tryVerifyContract } from "../../../utils/verifyContract";

export async function deployBridgedTokenBeacon(verbose = false) {
  const BridgedToken = await ethers.getContractFactory("BridgedToken");

  const l1TokenBeacon = await upgrades.deployBeacon(BridgedToken);
  await l1TokenBeacon.waitForDeployment();

  if (verbose) {
    console.log("L1TokenBeacon deployed, at address:", await l1TokenBeacon.getAddress());
  }

  const l2TokenBeacon = await upgrades.deployBeacon(BridgedToken);
  await l2TokenBeacon.waitForDeployment();
  if (verbose) {
    console.log("L2TokenBeacon deployed, at address:", await l2TokenBeacon.getAddress());
  }

  await tryStoreAddress(
    network.name,
    "l1TokenBeacon",
    await l1TokenBeacon.getAddress(),
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    l1TokenBeacon.deployTransaction.hash,
  );

  await tryStoreAddress(
    network.name,
    "l2TokenBeacon",
    await l2TokenBeacon.getAddress(),
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    l2TokenBeacon.deployTransaction.hash,
  );

  await tryVerifyContract(await l1TokenBeacon.getAddress());
  await tryVerifyContract(await l2TokenBeacon.getAddress());

  return { l1TokenBeacon, l2TokenBeacon };
}
