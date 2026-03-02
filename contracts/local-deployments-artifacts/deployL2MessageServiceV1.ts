import { ethers } from "ethers";
import * as dotenv from "dotenv";
import _json from "./dynamic-artifacts/L2MessageServiceV1.json" with { type: "json" };
const {
  contractName: L2MessageServiceContractName,
  abi: L2MessageServiceAbi,
  bytecode: L2MessageServiceBytecode,
} = _json;
import _json1 from "./static-artifacts/ProxyAdmin.json" with { type: "json" };
const { contractName: ProxyAdminContractName, abi: ProxyAdminAbi, bytecode: ProxyAdminBytecode } = _json1;
import _json2 from "./static-artifacts/TransparentUpgradeableProxy.json" with { type: "json" };
const { abi: TransparentUpgradeableProxyAbi, bytecode: TransparentUpgradeableProxyBytecode } = _json2;
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
  const messageServiceName = process.env.L2_MESSAGE_SERVICE_CONTRACT_NAME;

  if (!messageServiceName) {
    throw new Error("L2_MESSAGE_SERVICE_CONTRACT_NAME is required");
  }

  const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);
  const wallet = new ethers.Wallet(process.env.DEPLOYER_PRIVATE_KEY!, provider);
  let walletNonce;

  if (!process.env.L2_NONCE) {
    walletNonce = await wallet.getNonce();
  } else {
    walletNonce = parseInt(process.env.L2_NONCE);
  }

  const l2MessageServiceContractImplementationName = "L2MessageServiceImplementation";

  const [l2MessageServiceImplementation, proxyAdmin] = await Promise.all([
    deployContractFromArtifacts(
      l2MessageServiceContractImplementationName,
      L2MessageServiceAbi,
      L2MessageServiceBytecode,
      wallet,
      {
        nonce: walletNonce,
        maxFeePerGas: 7_200_000_000_000n,
        maxPriorityFeePerGas: 7_000_000_000_000n,
      },
    ),
    deployContractFromArtifacts(ProxyAdminContractName, ProxyAdminAbi, ProxyAdminBytecode, wallet, {
      nonce: walletNonce + 1,
      maxFeePerGas: 7_200_000_000_000n,
      maxPriorityFeePerGas: 7_000_000_000_000n,
    }),
  ]);

  const proxyAdminAddress = await proxyAdmin.getAddress();
  const l2MessageServiceImplementationAddress = await l2MessageServiceImplementation.getAddress();

  const pauseTypeRoles = getEnvVarOrDefault(
    "L2_MESSAGE_SERVICE_PAUSE_TYPES_ROLES",
    L2_MESSAGE_SERVICE_PAUSE_TYPES_ROLES,
  );
  const unpauseTypeRoles = getEnvVarOrDefault(
    "L2_MESSAGE_SERVICE_UNPAUSE_TYPES_ROLES",
    L2_MESSAGE_SERVICE_UNPAUSE_TYPES_ROLES,
  );
  const defaultRoleAddresses = generateRoleAssignments(L2_MESSAGE_SERVICE_ROLES, process.env.L2_SECURITY_COUNCIL!, [
    { role: L1_L2_MESSAGE_SETTER_ROLE, addresses: [process.env.L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER!] },
  ]);
  const roleAddresses = getEnvVarOrDefault("L2_MESSAGE_SERVICE_ROLE_ADDRESSES", defaultRoleAddresses);

  const initializer = getInitializerData(L2MessageServiceAbi, "initialize", [
    process.env.L2_MESSAGE_SERVICE_RATE_LIMIT_PERIOD,
    process.env.L2_MESSAGE_SERVICE_RATE_LIMIT_AMOUNT,
    process.env.L2_SECURITY_COUNCIL,
    roleAddresses,
    pauseTypeRoles,
    unpauseTypeRoles,
  ]);

  await deployContractFromArtifacts(
    L2MessageServiceContractName,
    TransparentUpgradeableProxyAbi,
    TransparentUpgradeableProxyBytecode,
    wallet,
    l2MessageServiceImplementationAddress,
    proxyAdminAddress,
    initializer,
    {
      maxFeePerGas: 7_200_000_000_000n,
      maxPriorityFeePerGas: 7_000_000_000_000n,
    },
  );
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
