import { ethers } from "ethers";
import fs from "fs";
import path from "path";
import * as dotenv from "dotenv";
import _json from "./dynamic-artifacts/ValidiumV1.json" with { type: "json" };
const { abi: ValidiumV1Abi, bytecode: ValidiumV1Bytecode } = _json;
import _json1 from "./static-artifacts/ProxyAdmin.json" with { type: "json" };
const { contractName: ProxyAdminContractName, abi: ProxyAdminAbi, bytecode: ProxyAdminBytecode } = _json1;
import _json2 from "./static-artifacts/TransparentUpgradeableProxy.json" with { type: "json" };
const { abi: TransparentUpgradeableProxyAbi, bytecode: TransparentUpgradeableProxyBytecode } = _json2;
import { getEnvVarOrDefault, getRequiredEnvVar } from "../common/helpers/environment.js";
import { deployContractFromArtifacts, getInitializerData } from "../common/helpers/deployments.js";
import { generateRoleAssignments } from "../common/helpers/roles.js";
import {
  VALIDIUM_PAUSE_TYPES_ROLES,
  VALIDIUM_ROLES,
  VALIDIUM_UNPAUSE_TYPES_ROLES,
  OPERATOR_ROLE,
} from "../common/constants/index.js";
import { get1559Fees } from "../scripts/utils.js";
import { ADDRESS_ZERO } from "../common/constants/general.js";

dotenv.config();

function findContractArtifacts(
  folderPath: string,
  contractName: string,
): { abi: ethers.InterfaceAbi; bytecode: ethers.BytesLike } {
  const files = fs.readdirSync(folderPath);

  const foundFile = files.find((file) => file === `${contractName}.json`);

  if (!foundFile) {
    // Throw an error if the file is not found
    throw new Error(`Contract "${contractName}" not found in folder "${folderPath}"`);
  }

  // Construct the full file path
  const filePath = path.join(folderPath, foundFile);

  // Read the file content
  const fileContent = fs.readFileSync(filePath, "utf-8").trim();
  const parsedContent = JSON.parse(fileContent);
  return parsedContent;
}

async function main() {
  const verifierName = getRequiredEnvVar("VERIFIER_CONTRACT_NAME");
  const validiumInitialStateRootHash = getRequiredEnvVar("INITIAL_L2_STATE_ROOT_HASH");
  const validiumInitialL2BlockNumber = getRequiredEnvVar("INITIAL_L2_BLOCK_NUMBER");
  const validiumSecurityCouncil = getRequiredEnvVar("L1_SECURITY_COUNCIL");
  const validiumOperators = getRequiredEnvVar("VALIDIUM_OPERATORS").split(",");
  const validiumRateLimitPeriodInSeconds = getRequiredEnvVar("VALIDIUM_RATE_LIMIT_PERIOD");
  const validiumRateLimitAmountInWei = getRequiredEnvVar("VALIDIUM_RATE_LIMIT_AMOUNT");
  const validiumGenesisTimestamp = getRequiredEnvVar("L2_GENESIS_TIMESTAMP");
  const validiumName = "Validium";
  const validiumImplementationName = "Validium";

  const pauseTypeRoles = getEnvVarOrDefault("VALIDIUM_PAUSE_TYPES_ROLES", VALIDIUM_PAUSE_TYPES_ROLES);
  const unpauseTypeRoles = getEnvVarOrDefault("VALIDIUM_UNPAUSE_TYPES_ROLES", VALIDIUM_UNPAUSE_TYPES_ROLES);
  const defaultRoleAddresses = generateRoleAssignments(VALIDIUM_ROLES, validiumSecurityCouncil, [
    { role: OPERATOR_ROLE, addresses: validiumOperators },
  ]);
  const roleAddresses = getEnvVarOrDefault("VALIDIUM_ROLE_ADDRESSES", defaultRoleAddresses);

  const verifierArtifacts = findContractArtifacts(path.join(import.meta.dirname, "./dynamic-artifacts"), verifierName);

  const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);

  const wallet = new ethers.Wallet(process.env.DEPLOYER_PRIVATE_KEY!, provider);

  const { gasPrice } = await get1559Fees(provider);

  let walletNonce;

  if (!process.env.L1_NONCE) {
    walletNonce = await wallet.getNonce();
  } else {
    walletNonce = parseInt(process.env.L1_NONCE);
  }

  const [verifier, validiumImplementation, proxyAdmin] = await Promise.all([
    deployContractFromArtifacts(verifierName, verifierArtifacts.abi, verifierArtifacts.bytecode, wallet, {
      nonce: walletNonce,
      gasPrice,
    }),
    deployContractFromArtifacts(validiumImplementationName, ValidiumV1Abi, ValidiumV1Bytecode, wallet, {
      nonce: walletNonce + 1,
      gasPrice,
    }),
    deployContractFromArtifacts(ProxyAdminContractName, ProxyAdminAbi, ProxyAdminBytecode, wallet, {
      nonce: walletNonce + 2,
      gasPrice,
    }),
  ]);

  const proxyAdminAddress = await proxyAdmin.getAddress();
  const verifierAddress = await verifier.getAddress();
  const validiumImplementationAddress = await validiumImplementation.getAddress();

  const initializer = getInitializerData(ValidiumV1Abi, "initialize", [
    {
      initialStateRootHash: validiumInitialStateRootHash,
      initialL2BlockNumber: validiumInitialL2BlockNumber,
      genesisTimestamp: validiumGenesisTimestamp,
      defaultVerifier: verifierAddress,
      rateLimitPeriodInSeconds: validiumRateLimitPeriodInSeconds,
      rateLimitAmountInWei: validiumRateLimitAmountInWei,
      roleAddresses,
      pauseTypeRoles,
      unpauseTypeRoles,
      defaultAdmin: validiumSecurityCouncil,
      shnarfProvider: ADDRESS_ZERO,
    },
  ]);

  await deployContractFromArtifacts(
    validiumName,
    TransparentUpgradeableProxyAbi,
    TransparentUpgradeableProxyBytecode,
    wallet,
    validiumImplementationAddress,
    proxyAdminAddress,
    initializer,
    { gasPrice },
  );
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
