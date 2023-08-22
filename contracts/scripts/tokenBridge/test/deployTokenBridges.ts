import { ethers, upgrades } from "hardhat";

import { deployBridgedTokenBeacon } from "./deployBridgedTokenBeacon";
import { SupportedChainIds } from "../supportedNetworks";

export async function deployTokenBridge(messageServiceAddress: string, verbose = false) {
  const [owner] = await ethers.getSigners();
  const chainIds = [SupportedChainIds.GOERLI, SupportedChainIds.LINEA_TESTNET];

  // Deploy beacon for bridged tokens
  const tokenBeacons = await deployBridgedTokenBeacon(verbose);

  // Deploying TokenBridges
  const TokenBridgeFactory = await ethers.getContractFactory("TokenBridge");

  const l1TokenBridge = await upgrades.deployProxy(TokenBridgeFactory, [
    owner.address,
    messageServiceAddress,
    tokenBeacons.l1TokenBeacon.address,
    chainIds[0],
    chainIds[1],
    [], // Reseved Addresses
  ]);
  await l1TokenBridge.deployed();
  if (verbose) {
    console.log("L1TokenBridge deployed, at address:", l1TokenBridge.address);
  }

  const l2TokenBridge = await upgrades.deployProxy(TokenBridgeFactory, [
    owner.address,
    messageServiceAddress,
    tokenBeacons.l2TokenBeacon.address,
    chainIds[1],
    chainIds[0],
    [], // Reseved Addresses
  ]);
  await l2TokenBridge.deployed();
  if (verbose) {
    console.log("L2TokenBridge deployed, at address:", l2TokenBridge.address);
  }

  // Setting reciprocal addresses of TokenBridges
  await l1TokenBridge.setRemoteTokenBridge(l2TokenBridge.address);
  await l2TokenBridge.setRemoteTokenBridge(l1TokenBridge.address);
  if (verbose) {
    console.log("Reciprocal addresses of TokenBridges set");
  }

  if (verbose) {
    console.log("Deployment finished");
  }

  return { l1TokenBridge, l2TokenBridge, chainIds, ...tokenBeacons };
}

export async function deployTokenBridgeWithMockMessaging(verbose = false) {
  const MessageServiceFactory = await ethers.getContractFactory("MockMessageService");

  // Deploying mock messaging service
  const messageService = await MessageServiceFactory.deploy();
  await messageService.deployed();

  const deploymentVars = await deployTokenBridge(messageService.address, verbose);
  return { messageService, ...deploymentVars };
}
