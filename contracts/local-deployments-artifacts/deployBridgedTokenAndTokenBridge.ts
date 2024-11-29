import { abi as ProxyAdminAbi, bytecode as ProxyAdminBytecode } from "./static-artifacts/ProxyAdmin.json";
import { abi as BridgedTokenAbi, bytecode as BridgedTokenBytecode } from "./dynamic-artifacts/BridgedToken.json";
import { abi as TokenBridgeAbi, bytecode as TokenBridgeBytecode } from "./dynamic-artifacts/TokenBridge.json";
import {
  abi as UpgradeableBeaconAbi,
  bytecode as UpgradeableBeaconBytecode,
} from "./static-artifacts/UpgradeableBeacon.json";

import {
  abi as TransparentUpgradeableProxyAbi,
  bytecode as TransparentUpgradeableProxyBytecode,
} from "./static-artifacts/TransparentUpgradeableProxy.json";
import { getEnvVarOrDefault, getRequiredEnvVar } from "../common/helpers/environment";
import { generateRoleAssignments } from "../common/helpers/roles";
import {
  TOKEN_BRIDGE_PAUSE_TYPES_ROLES,
  TOKEN_BRIDGE_ROLES,
  TOKEN_BRIDGE_UNPAUSE_TYPES_ROLES,
} from "contracts/common/constants";
import { ethers } from "ethers";
import { deployContractFromArtifacts, getInitializerData } from "../common/helpers/deployments";

async function main() {
  const ORDERED_NONCE_POST_L2MESSAGESERVICE = 3;
  const ORDERED_NONCE_POST_LINEAROLLUP = 4;
  const bridgedTokenName = "BridgedToken";
  const tokenBridgeName = "TokenBridge";

  const l2MessageServiceAddress = process.env.L2_MESSAGE_SERVICE_ADDRESS;
  const lineaRollupAddress = process.env.LINEA_ROLLUP_ADDRESS;

  const remoteChainId = getRequiredEnvVar("REMOTE_CHAIN_ID");
  const tokenBridgeSecurityCouncil = getRequiredEnvVar("TOKEN_BRIDGE_SECURITY_COUNCIL");

  const pauseTypeRoles = getEnvVarOrDefault("TOKEN_BRIDGE_PAUSE_TYPES_ROLES", TOKEN_BRIDGE_PAUSE_TYPES_ROLES);
  const unpauseTypeRoles = getEnvVarOrDefault("TOKEN_BRIDGE_UNPAUSE_TYPES_ROLES", TOKEN_BRIDGE_UNPAUSE_TYPES_ROLES);
  const defaultRoleAddresses = generateRoleAssignments(TOKEN_BRIDGE_ROLES, tokenBridgeSecurityCouncil, []);
  const roleAddresses = getEnvVarOrDefault("TOKEN_BRIDGE_ROLE_ADDRESSES", defaultRoleAddresses);
  const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);
  const wallet = new ethers.Wallet(process.env.PRIVATE_KEY!, provider);

  let walletNonce;

  if (process.env.TOKEN_BRIDGE_L1 === "true") {
    if (process.env.L1_NONCE === undefined) {
      walletNonce = await wallet.getNonce();
    } else {
      walletNonce = parseInt(process.env.L1_NONCE) + ORDERED_NONCE_POST_LINEAROLLUP;
    }
  } else {
    if (process.env.L2_NONCE === undefined) {
      walletNonce = await wallet.getNonce();
    } else {
      walletNonce = parseInt(process.env.L2_NONCE) + ORDERED_NONCE_POST_L2MESSAGESERVICE;
    }
  }

  const [bridgedToken, tokenBridgeImplementation, proxyAdmin] = await Promise.all([
    deployContractFromArtifacts(BridgedTokenAbi, BridgedTokenBytecode, wallet, { nonce: walletNonce }),
    deployContractFromArtifacts(TokenBridgeAbi, TokenBridgeBytecode, wallet, { nonce: walletNonce + 1 }),
    deployContractFromArtifacts(ProxyAdminAbi, ProxyAdminBytecode, wallet, { nonce: walletNonce + 2 }),
  ]);

  const bridgedTokenAddress = await bridgedToken.getAddress();
  const tokenBridgeImplementationAddress = await tokenBridgeImplementation.getAddress();
  const proxyAdminAddress = await proxyAdmin.getAddress();

  const chainId = (await provider.getNetwork()).chainId;

  console.log(`${bridgedTokenName} contract deployed at ${bridgedTokenAddress}`);
  console.log(`${tokenBridgeName} Implementation contract deployed at ${tokenBridgeImplementationAddress}`);
  console.log(`L1 ProxyAdmin deployed: address=${proxyAdminAddress}`);
  console.log(`Deploying UpgradeableBeacon: chainId=${chainId} bridgedTokenAddress=${bridgedTokenAddress}`);

  const beaconProxy = await deployContractFromArtifacts(
    UpgradeableBeaconAbi,
    UpgradeableBeaconBytecode,
    wallet,
    bridgedTokenAddress,
  );

  const beaconProxyAddress = await beaconProxy.getAddress();

  let deployingChainMessageService = l2MessageServiceAddress;
  let reservedAddresses = process.env.L2_RESERVED_TOKEN_ADDRESSES
    ? process.env.L2_RESERVED_TOKEN_ADDRESSES.split(",")
    : [];

  if (process.env.TOKEN_BRIDGE_L1 === "true") {
    console.log(
      `TOKEN_BRIDGE_L1=${process.env.TOKEN_BRIDGE_L1}. Deploying TokenBridge on L1, using L1_RESERVED_TOKEN_ADDRESSES environment variable`,
    );
    deployingChainMessageService = lineaRollupAddress;
    reservedAddresses = process.env.L1_RESERVED_TOKEN_ADDRESSES
      ? process.env.L1_RESERVED_TOKEN_ADDRESSES.split(",")
      : [];
  } else {
    console.log(
      `TOKEN_BRIDGE_L1=${process.env.TOKEN_BRIDGE_L1}. Deploying TokenBridge on L2, using L2_RESERVED_TOKEN_ADDRESSES environment variable`,
    );
  }

  const initializer = getInitializerData(TokenBridgeAbi, "initialize", [
    {
      defaultAdmin: tokenBridgeSecurityCouncil,
      messageService: deployingChainMessageService,
      tokenBeacon: beaconProxyAddress,
      sourceChainId: chainId,
      targetChainId: remoteChainId,
      reservedTokens: reservedAddresses,
      roleAddresses,
      pauseTypeRoles,
      unpauseTypeRoles,
    },
  ]);

  const proxyContract = await deployContractFromArtifacts(
    TransparentUpgradeableProxyAbi,
    TransparentUpgradeableProxyBytecode,
    wallet,
    tokenBridgeImplementationAddress,
    proxyAdminAddress,
    initializer,
  );

  const proxyContractAddress = await proxyContract.getAddress();
  const txReceipt = await proxyContract.deploymentTransaction()?.wait();

  if (!txReceipt) {
    throw "Contract deployment transaction receipt not found.";
  }

  console.log(
    `${tokenBridgeName} deployed: chainId=${chainId} address=${proxyContractAddress} blockNumber=${txReceipt.blockNumber}`,
  );
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
