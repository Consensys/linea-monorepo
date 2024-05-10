import { ethers, network, upgrades } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { validateDeployBranchAndTags } from "../utils/auditedDeployVerifier";
import { getDeployedContractAddress, tryStoreAddress, tryStoreProxyAdminAddress } from "../utils/storeAddress";
import { tryVerifyContract } from "../utils/verifyContract";
import { requireEnv } from "../scripts/hardhat/utils";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments } = hre;
  validateDeployBranchAndTags(hre.network.name);

  const contractName = "TokenBridge";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);

  const L2MessageServiceName = "L2MessageService";
  const LineaRollupName = "LineaRollup";
  let l2MessageServiceAddress = process.env.L2_MESSAGE_SERVICE_ADDRESS;
  let LineaRollupAddress = process.env.LINEA_ROLLUP_ADDRESS;
  const remoteChainId = requireEnv("REMOTE_CHAIN_ID");

  const [owner] = await ethers.getSigners();
  const chainId = (await ethers.provider.getNetwork()).chainId;

  console.log(`Current network's chainId is ${chainId}. Remote (target) network's chainId is ${remoteChainId}`);

  if (l2MessageServiceAddress === undefined) {
    l2MessageServiceAddress = await getDeployedContractAddress(L2MessageServiceName, deployments);
  }

  if (LineaRollupAddress === undefined) {
    LineaRollupAddress = await getDeployedContractAddress(LineaRollupName, deployments);
  }

  if (existingContractAddress === undefined) {
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
    deployingChainMessageService = LineaRollupAddress;
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
    owner.address,
    deployingChainMessageService,
    bridgedTokenAddress,
    chainId,
    remoteChainId,
    reservedAddresses,
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

  console.log(`TokenBridge deployed on ${network.name}, at address: ${tokenBridgeAddress}`);

  await tryVerifyContract(tokenBridgeAddress);
};
export default func;
func.tags = ["TokenBridge"];
