import { ContractFactory, ethers } from "ethers";
import * as dotenv from "dotenv";
import {
  abi as L2MessageServiceAbi,
  bytecode as L2MessageServiceBytecode,
} from "./shared-artifacts/L2MessageService.json";
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

dotenv.config();

async function deployContract(
  abi: ethers.Interface | ethers.InterfaceAbi,
  bytecode: ethers.BytesLike,
  wallet: ethers.ContractRunner,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ...args: ethers.ContractMethodArgs<any[]>
) {
  const factory = new ethers.ContractFactory(abi, bytecode, wallet);
  const contract = await factory.deploy(...args);
  await contract.waitForDeployment();
  return contract.getAddress();
}

async function main() {
  const messageServiceName = process.env.MESSAGE_SERVICE_CONTRACT_NAME;

  if (!messageServiceName) {
    throw new Error("MESSAGE_SERVICE_CONTRACT_NAME is required");
  }

  const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);
  const wallet = new ethers.Wallet(process.env.PRIVATE_KEY!, provider);
  const walletNonce = await wallet.getNonce();

  // Deploy the implementation contract
  const [implementationAddress, proxyAdminAddress] = await Promise.all([
    deployContract(L2MessageServiceAbi, L2MessageServiceBytecode, wallet, { nonce: walletNonce }),
    deployContract(ProxyAdminAbi, ProxyAdminBytecode, wallet, { nonce: walletNonce + 1 }),
  ]);

  // Deploy the proxy admin contract
  console.log(`Proxy admin contract deployed at ${proxyAdminAddress}`);
  console.log(`${messageServiceName} Implementation contract deployed at ${implementationAddress}`);

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
  // Prepare the initializer data
  const initializer = new ContractFactory(L2MessageServiceAbi, L2MessageServiceBytecode).interface.encodeFunctionData(
    "initialize",
    [
      process.env.L2MSGSERVICE_RATE_LIMIT_PERIOD,
      process.env.L2MSGSERVICE_RATE_LIMIT_AMOUNT,
      process.env.L2MSGSERVICE_SECURITY_COUNCIL,
      roleAddresses,
      pauseTypeRoles,
      unpauseTypeRoles,
    ],
  );

  // Deploy the proxy contract

  const proxyAddress = await deployContract(
    TransparentUpgradeableProxyAbi,
    TransparentUpgradeableProxyBytecode,
    wallet,
    implementationAddress,
    proxyAdminAddress,
    initializer,
  );
  console.log(`${messageServiceName} Proxy contract deployed at ${proxyAddress}`);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
