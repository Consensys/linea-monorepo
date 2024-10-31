import { ethers } from "ethers";
import * as dotenv from "dotenv";
import {
  abi as L2MessageServiceAbi,
  bytecode as L2MessageServiceBytecode,
} from "./dynamic-artifacts/L2MessageService.json";
import { abi as ProxyAdminAbi, bytecode as ProxyAdminBytecode } from "./static-artifacts/ProxyAdmin.json";
import {
  abi as TransparentUpgradeableProxyAbi,
  bytecode as TransparentUpgradeableProxyBytecode,
} from "./static-artifacts/TransparentUpgradeableProxy.json";
import { getEnvVarOrDefault } from "../common/helpers/environment";
import {
  L1_L2_MESSAGE_SETTER_ROLE,
  L2_MESSAGE_SERVICE_PAUSE_TYPES_ROLES,
  L2_MESSAGE_SERVICE_ROLES,
  L2_MESSAGE_SERVICE_UNPAUSE_TYPES_ROLES,
} from "../common/constants";
import { generateRoleAssignments } from "../common/helpers/roles";
import { deployContractFromArtifacts, getInitializerData } from "../common/helpers/deployments";

dotenv.config();

async function main() {
  const messageServiceName = process.env.MESSAGE_SERVICE_CONTRACT_NAME;

  if (!messageServiceName) {
    throw new Error("MESSAGE_SERVICE_CONTRACT_NAME is required");
  }

  const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);
  const wallet = new ethers.Wallet(process.env.PRIVATE_KEY!, provider);
  let walletNonce;

  if (!process.env.L2_NONCE) {
    walletNonce = await wallet.getNonce();
  } else {
    walletNonce = parseInt(process.env.L2_NONCE);
  }

  const [l2MessageServiceImplementation, proxyAdmin] = await Promise.all([
    deployContractFromArtifacts(L2MessageServiceAbi, L2MessageServiceBytecode, wallet, { nonce: walletNonce }),
    deployContractFromArtifacts(ProxyAdminAbi, ProxyAdminBytecode, wallet, { nonce: walletNonce + 1 }),
  ]);

  const proxyAdminAddress = await proxyAdmin.getAddress();
  const l2MessageServiceImplementationAddress = await l2MessageServiceImplementation.getAddress();

  console.log(`L2 ProxyAdmin deployed: address=${proxyAdminAddress}`);

  const pauseTypeRoles = getEnvVarOrDefault("L2MSGSERVICE_PAUSE_TYPE_ROLES", L2_MESSAGE_SERVICE_PAUSE_TYPES_ROLES);
  const unpauseTypeRoles = getEnvVarOrDefault(
    "L2MSGSERVICE_UNPAUSE_TYPE_ROLES",
    L2_MESSAGE_SERVICE_UNPAUSE_TYPES_ROLES,
  );
  const defaultRoleAddresses = generateRoleAssignments(
    L2_MESSAGE_SERVICE_ROLES,
    process.env.L2MSGSERVICE_SECURITY_COUNCIL!,
    [{ role: L1_L2_MESSAGE_SETTER_ROLE, addresses: [process.env.L2MSGSERVICE_L1L2_MESSAGE_SETTER!] }],
  );
  const roleAddresses = getEnvVarOrDefault("L2MSGSERVICE_ROLE_ADDRESSES", defaultRoleAddresses);

  const initializer = getInitializerData(L2MessageServiceAbi, "initialize", [
    process.env.L2MSGSERVICE_RATE_LIMIT_PERIOD,
    process.env.L2MSGSERVICE_RATE_LIMIT_AMOUNT,
    process.env.L2MSGSERVICE_SECURITY_COUNCIL,
    roleAddresses,
    pauseTypeRoles,
    unpauseTypeRoles,
  ]);

  const proxyContract = await deployContractFromArtifacts(
    TransparentUpgradeableProxyAbi,
    TransparentUpgradeableProxyBytecode,
    wallet,
    l2MessageServiceImplementationAddress,
    proxyAdminAddress,
    initializer,
  );

  const proxyContractAddress = await proxyContract.getAddress();
  const txReceipt = await proxyContract.deploymentTransaction()?.wait();

  if (!txReceipt) {
    throw "Contract deployment transaction receipt not found.";
  }

  console.log(`${messageServiceName} deployed: address=${proxyContractAddress} blockNumber=${txReceipt.blockNumber}`);
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
