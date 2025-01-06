import {
  contractName as ProxyAdminContractName,
  abi as ProxyAdminAbi,
  bytecode as ProxyAdminBytecode,
} from "./static-artifacts/ProxyAdmin.json";
import {
  contractName as BridgedTokenContractName,
  abi as BridgedTokenAbi,
  bytecode as BridgedTokenBytecode,
} from "./dynamic-artifacts/BridgedToken.json";
import {
  contractName as TokenBridgeContractName,
  abi as TokenBridgeAbi,
  bytecode as TokenBridgeBytecode,
} from "./dynamic-artifacts/TokenBridge.json";
import {
  contractName as UpgradeableBeaconContractName,
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

  const l2MessageServiceAddress = process.env.L2MESSAGESERVICE_ADDRESS;
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

  const tokenBridgeContractImplementationName = "tokenBridgeContractImplementation";

  const [bridgedToken, tokenBridgeImplementation, proxyAdmin] = await Promise.all([
    deployContractFromArtifacts(BridgedTokenContractName, BridgedTokenAbi, BridgedTokenBytecode, wallet, {
      nonce: walletNonce,
    }),
    deployContractFromArtifacts(tokenBridgeContractImplementationName, TokenBridgeAbi, TokenBridgeBytecode, wallet, {
      nonce: walletNonce + 1,
    }),
    deployContractFromArtifacts(ProxyAdminContractName, ProxyAdminAbi, ProxyAdminBytecode, wallet, {
      nonce: walletNonce + 2,
    }),
  ]);

  const bridgedTokenAddress = await bridgedToken.getAddress();
  const tokenBridgeImplementationAddress = await tokenBridgeImplementation.getAddress();
  const proxyAdminAddress = await proxyAdmin.getAddress();

  const chainId = (await provider.getNetwork()).chainId;

  console.log(`Deploying UpgradeableBeacon: chainId=${chainId} bridgedTokenAddress=${bridgedTokenAddress}`);

  const beaconProxy = await deployContractFromArtifacts(
    UpgradeableBeaconContractName,
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

  await deployContractFromArtifacts(
    TokenBridgeContractName,
    TransparentUpgradeableProxyAbi,
    TransparentUpgradeableProxyBytecode,
    wallet,
    tokenBridgeImplementationAddress,
    proxyAdminAddress,
    initializer,
  );
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
