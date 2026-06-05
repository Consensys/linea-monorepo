import { upgrades as createUpgrades } from "@openzeppelin/hardhat-upgrades";
import { TOKEN_BRIDGE_ROLES } from "contracts/common/constants";
import { generateRoleAssignments } from "contracts/common/helpers";
import { PauseTypeRole, RoleAddress } from "contracts/test/hardhat/common/types";
import hre, { network as hardhatNetwork } from "hardhat";

import { deployBridgedTokenBeacon } from "./deployBridgedTokenBeacon";
import { SupportedChainIds } from "../../../common/supportedNetworks";
import { pauseTypeRoles, unpauseTypeRoles } from "../../../test/hardhat/common/constants";

import type { TokenBridge } from "../../../typechain-types";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;
const upgrades = await createUpgrades(hre, hardhatConnection);

export async function deployTokenBridge(messageServiceAddress: string, verbose = false) {
  const [owner] = await ethers.getSigners();
  const chainIds = [SupportedChainIds.SEPOLIA, SupportedChainIds.LINEA_TESTNET];

  const roleAddresses = generateRoleAssignments(TOKEN_BRIDGE_ROLES, owner.address, []);

  // Deploy beacon for bridged tokens
  const tokenBeacons = await deployBridgedTokenBeacon(verbose);

  // Deploying TokenBridges
  const TokenBridgeFactory = await ethers.getContractFactory("TokenBridge");

  await upgrades.deployImplementation(TokenBridgeFactory);

  // deployProxy deploys the transparent proxy in one external deployer transaction. OZ v5 creates ProxyAdmin
  // inside the proxy constructor, so it does not consume an additional deployer nonce.
  const l2TokenBridgeNonceOffset = 1;

  const l1TokenBridgeInitializationData: TokenBridgeInitializationData = {
    defaultAdmin: owner.address,
    messageService: messageServiceAddress,
    tokenBeacon: await tokenBeacons.l1TokenBeacon.getAddress(),
    sourceChainId: chainIds[0],
    targetChainId: chainIds[1],
    reservedTokens: [],
    remoteSender: ethers.getCreateAddress({
      from: await owner.getAddress(),
      nonce: l2TokenBridgeNonceOffset + (await owner.getNonce()),
    }), // Counterfactual address of l2TokenBridge
    roleAddresses: roleAddresses,
    pauseTypeRoles: pauseTypeRoles.map((role) => ({ pauseType: role.pauseType.toString(), role: role.role })),
    unpauseTypeRoles: unpauseTypeRoles.map((role) => ({ pauseType: role.pauseType.toString(), role: role.role })),
  };

  const l1TokenBridge = (await upgrades.deployProxy(TokenBridgeFactory, [l1TokenBridgeInitializationData], {
    initialOwner: owner.address,
  })) as unknown as TokenBridge;
  await l1TokenBridge.waitForDeployment();
  if (verbose) {
    console.log("L1TokenBridge deployed, at address:", await l1TokenBridge.getAddress());
  }

  const l2TokenBridgeInitializationData: TokenBridgeInitializationData = {
    defaultAdmin: owner.address,
    messageService: messageServiceAddress,
    tokenBeacon: await tokenBeacons.l2TokenBeacon.getAddress(),
    sourceChainId: chainIds[1],
    targetChainId: chainIds[0],
    reservedTokens: [],
    remoteSender: await l1TokenBridge.getAddress(),
    roleAddresses: roleAddresses,
    pauseTypeRoles: pauseTypeRoles.map((role) => ({ pauseType: role.pauseType.toString(), role: role.role })),
    unpauseTypeRoles: unpauseTypeRoles.map((role) => ({ pauseType: role.pauseType.toString(), role: role.role })),
  };

  const l2TokenBridge = (await upgrades.deployProxy(TokenBridgeFactory, [l2TokenBridgeInitializationData], {
    initialOwner: owner.address,
  })) as unknown as TokenBridge;
  await l2TokenBridge.waitForDeployment();
  if (verbose) {
    console.log("L2TokenBridge deployed, at address:", await l2TokenBridge.getAddress());
  }

  if (verbose) {
    console.log("Deployment finished");
  }

  return {
    l1TokenBridge,
    l2TokenBridge,
    chainIds,
    ...tokenBeacons,
    l1TokenBridgeInitializationData,
    l2TokenBridgeInitializationData,
  };
}

export type TokenBridgeInitializationData = {
  defaultAdmin: string;
  messageService: string;
  tokenBeacon: string;
  sourceChainId: number;
  targetChainId: number;
  reservedTokens: string[];
  remoteSender: string;
  roleAddresses: RoleAddress[];
  pauseTypeRoles: PauseTypeRole[];
  unpauseTypeRoles: PauseTypeRole[];
};

export async function deployTokenBridgeWithMockMessaging(verbose = false) {
  const MessageServiceFactory = await ethers.getContractFactory("MockMessageService");

  // Deploying mock messaging service
  const messageService = await MessageServiceFactory.deploy();
  await messageService.waitForDeployment();

  const deploymentVars = await deployTokenBridge(await messageService.getAddress(), verbose);
  return { messageService, ...deploymentVars };
}
