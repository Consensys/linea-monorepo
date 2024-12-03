import { ethers, network, upgrades } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import {
  TOKEN_BRIDGE_PAUSE_TYPES_ROLES,
  TOKEN_BRIDGE_ROLES,
  TOKEN_BRIDGE_UNPAUSE_TYPES_ROLES,
} from "../common/constants";
import {
  generateRoleAssignments,
  getEnvVarOrDefault,
  tryVerifyContract,
  getDeployedContractAddress,
  tryStoreAddress,
  tryStoreProxyAdminAddress,
  getRequiredEnvVar,
} from "../common/helpers";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments } = hre;

  const contractName = "TokenBridge";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);

  const l2MessageServiceName = "L2MessageService";
  const lineaRollupName = "LineaRollup";
  let l2MessageServiceAddress = process.env.L2_MESSAGE_SERVICE_ADDRESS;
  let lineaRollupAddress = process.env.LINEA_ROLLUP_ADDRESS;
  const remoteChainId = getRequiredEnvVar("REMOTE_CHAIN_ID");
  const tokenBridgeSecurityCouncil = getRequiredEnvVar("TOKEN_BRIDGE_SECURITY_COUNCIL");

  const pauseTypeRoles = getEnvVarOrDefault("TOKEN_BRIDGE_PAUSE_TYPES_ROLES", TOKEN_BRIDGE_PAUSE_TYPES_ROLES);
  const unpauseTypeRoles = getEnvVarOrDefault("TOKEN_BRIDGE_UNPAUSE_TYPES_ROLES", TOKEN_BRIDGE_UNPAUSE_TYPES_ROLES);
  const defaultRoleAddresses = generateRoleAssignments(TOKEN_BRIDGE_ROLES, tokenBridgeSecurityCouncil, []);
  const roleAddresses = getEnvVarOrDefault("TOKEN_BRIDGE_ROLE_ADDRESSES", defaultRoleAddresses);

  const chainId = (await ethers.provider.getNetwork()).chainId;

  console.log(`Current network's chainId is ${chainId}. Remote (target) network's chainId is ${remoteChainId}`);

  if (!l2MessageServiceAddress) {
    l2MessageServiceAddress = await getDeployedContractAddress(l2MessageServiceName, deployments);
  }

  if (!lineaRollupAddress) {
    lineaRollupAddress = await getDeployedContractAddress(lineaRollupName, deployments);
  }

  if (!existingContractAddress) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }

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

  let bridgedTokenAddress = await getDeployedContractAddress("BridgedToken", deployments);
  if (bridgedTokenAddress === undefined) {
    console.log(`Using environment variable for BridgedToken , ${process.env.BRIDGED_TOKEN_ADDRESS}`);
    if (process.env.BRIDGED_TOKEN_ADDRESS !== undefined) {
      bridgedTokenAddress = process.env.BRIDGED_TOKEN_ADDRESS;
    } else {
      throw "Missing BRIDGED_TOKEN_ADDRESS environment variable.";
    }
  }
  // Deploying TokenBridge
  const TokenBridgeFactory = await ethers.getContractFactory(contractName);

  const tokenBridge = await upgrades.deployProxy(TokenBridgeFactory, [
    {
      defaultAdmin: tokenBridgeSecurityCouncil,
      messageService: deployingChainMessageService,
      tokenBeacon: bridgedTokenAddress,
      sourceChainId: chainId,
      targetChainId: remoteChainId,
      reservedTokens: reservedAddresses,
      roleAddresses,
      pauseTypeRoles,
      unpauseTypeRoles,
    },
  ]);

  await tokenBridge.waitForDeployment();
  const tokenBridgeAddress = await tokenBridge.getAddress();

  const deployTx = tokenBridge.deploymentTransaction();
  if (!deployTx) {
    throw "Contract deployment transaction receipt not found.";
  }

  await tryStoreAddress(network.name, contractName, tokenBridgeAddress, deployTx.hash);

  const proxyAdminAddress = await upgrades.erc1967.getAdminAddress(tokenBridgeAddress);

  await tryStoreProxyAdminAddress(network.name, contractName, proxyAdminAddress);

  if (process.env.TOKEN_BRIDGE_L1 === "true") {
    console.log(`L1 TokenBridge deployed on ${network.name}, at address: ${tokenBridgeAddress}`);
  } else {
    console.log(`L2 TokenBridge deployed on ${network.name}, at address: ${tokenBridgeAddress}`);
  }
  await tryVerifyContract(tokenBridgeAddress);
};
export default func;
func.tags = ["TokenBridge"];
