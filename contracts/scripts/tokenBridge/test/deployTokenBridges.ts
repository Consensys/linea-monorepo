import { ethers } from "../../../test/hardhat/common/connection.js";

import type { TokenBridge } from "../../../typechain-types";
import { SupportedChainIds } from "../../../common/supportedNetworks";
import { deployBridgedTokenBeacon } from "./deployBridgedTokenBeacon";
import { pauseTypeRoles, unpauseTypeRoles } from "../../../test/hardhat/common/constants";
import { generateRoleAssignments } from "contracts/common/helpers";
import { TOKEN_BRIDGE_ROLES } from "contracts/common/constants";
import { deployTransparentProxy } from "../../../test/hardhat/common/deployment";

export async function deployTokenBridge(messageServiceAddress: string, verbose = false) {
  const [owner] = await ethers.getSigners();
  const chainIds = [SupportedChainIds.SEPOLIA, SupportedChainIds.LINEA_TESTNET];

  const roleAddresses = generateRoleAssignments(TOKEN_BRIDGE_ROLES, owner.address, []);

  const tokenBeacons = await deployBridgedTokenBeacon(verbose);

  const TokenBridgeFactory = await ethers.getContractFactory("TokenBridge");

  const l1TokenBridge = (await deployTransparentProxy(TokenBridgeFactory, [
    {
      defaultAdmin: owner.address,
      messageService: messageServiceAddress,
      tokenBeacon: await tokenBeacons.l1TokenBeacon.getAddress(),
      sourceChainId: chainIds[0],
      targetChainId: chainIds[1],
      reservedTokens: [],
      remoteSender: ethers.getCreateAddress({ from: await owner.getAddress(), nonce: 1 + (await owner.getNonce()) }),
      roleAddresses: roleAddresses,
      pauseTypeRoles: pauseTypeRoles,
      unpauseTypeRoles: unpauseTypeRoles,
    },
  ])) as unknown as TokenBridge;
  if (verbose) {
    console.log("L1TokenBridge deployed, at address:", await l1TokenBridge.getAddress());
  }

  const l2TokenBridge = (await deployTransparentProxy(TokenBridgeFactory, [
    {
      defaultAdmin: owner.address,
      messageService: messageServiceAddress,
      tokenBeacon: await tokenBeacons.l2TokenBeacon.getAddress(),
      sourceChainId: chainIds[1],
      targetChainId: chainIds[0],
      reservedTokens: [],
      remoteSender: await l1TokenBridge.getAddress(),
      roleAddresses: roleAddresses,
      pauseTypeRoles: pauseTypeRoles,
      unpauseTypeRoles: unpauseTypeRoles,
    },
  ])) as unknown as TokenBridge;
  if (verbose) {
    console.log("L2TokenBridge deployed, at address:", await l2TokenBridge.getAddress());
  }

  if (verbose) {
    console.log("Deployment finished");
  }

  return { l1TokenBridge, l2TokenBridge, chainIds, ...tokenBeacons };
}

export async function deployTokenBridgeWithMockMessaging(verbose = false) {
  const MessageServiceFactory = await ethers.getContractFactory("MockMessageService");

  const messageService = await MessageServiceFactory.deploy();
  await messageService.waitForDeployment();

  const deploymentVars = await deployTokenBridge(await messageService.getAddress(), verbose);
  return { messageService, ...deploymentVars };
}
