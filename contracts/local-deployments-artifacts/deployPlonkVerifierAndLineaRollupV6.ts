import { ethers } from "ethers";
import fs from "fs";
import path from "path";
import * as dotenv from "dotenv";
import { abi as LineaRollupV6Abi, bytecode as LineaRollupV6Bytecode } from "./dynamic-artifacts/LineaRollupV6.json";
import { abi as ProxyAdminAbi, bytecode as ProxyAdminBytecode } from "./static-artifacts/ProxyAdmin.json";
import {
  abi as TransparentUpgradeableProxyAbi,
  bytecode as TransparentUpgradeableProxyBytecode,
} from "./static-artifacts/TransparentUpgradeableProxy.json";
import { getEnvVarOrDefault, getRequiredEnvVar } from "../common/helpers/environment";
import { deployContractFromArtifacts, getInitializerData } from "../common/helpers/deployments";
import { generateRoleAssignments } from "../common/helpers/roles";
import {
  LINEA_ROLLUP_PAUSE_TYPES_ROLES,
  LINEA_ROLLUP_ROLES,
  LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
  OPERATOR_ROLE,
} from "../common/constants";
import { get1559Fees } from "../scripts/utils";

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
  const lineaRollupInitialStateRootHash = getRequiredEnvVar("LINEA_ROLLUP_INITIAL_STATE_ROOT_HASH");
  const lineaRollupInitialL2BlockNumber = getRequiredEnvVar("LINEA_ROLLUP_INITIAL_L2_BLOCK_NUMBER");
  const lineaRollupSecurityCouncil = getRequiredEnvVar("LINEA_ROLLUP_SECURITY_COUNCIL");
  const lineaRollupOperators = getRequiredEnvVar("LINEA_ROLLUP_OPERATORS").split(",");
  const lineaRollupRateLimitPeriodInSeconds = getRequiredEnvVar("LINEA_ROLLUP_RATE_LIMIT_PERIOD");
  const lineaRollupRateLimitAmountInWei = getRequiredEnvVar("LINEA_ROLLUP_RATE_LIMIT_AMOUNT");
  const lineaRollupGenesisTimestamp = getRequiredEnvVar("LINEA_ROLLUP_GENESIS_TIMESTAMP");
  const multiCallAddress = "0xcA11bde05977b3631167028862bE2a173976CA11";
  const lineaRollupName = "LineaRollupV6";

  const pauseTypeRoles = getEnvVarOrDefault("LINEA_ROLLUP_PAUSE_TYPE_ROLES", LINEA_ROLLUP_PAUSE_TYPES_ROLES);
  const unpauseTypeRoles = getEnvVarOrDefault("LINEA_ROLLUP_UNPAUSE_TYPE_ROLES", LINEA_ROLLUP_UNPAUSE_TYPES_ROLES);
  const defaultRoleAddresses = generateRoleAssignments(LINEA_ROLLUP_ROLES, lineaRollupSecurityCouncil, [
    { role: OPERATOR_ROLE, addresses: lineaRollupOperators },
  ]);
  const roleAddresses = getEnvVarOrDefault("LINEA_ROLLUP_ROLE_ADDRESSES", defaultRoleAddresses);

  const verifierArtifacts = findContractArtifacts(path.join(__dirname, "./dynamic-artifacts"), verifierName);

  const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);

  const wallet = new ethers.Wallet(process.env.PRIVATE_KEY!, provider);

  const { gasPrice } = await get1559Fees(provider);

  let walletNonce;

  if (!process.env.L1_NONCE) {
    walletNonce = await wallet.getNonce();
  } else {
    walletNonce = parseInt(process.env.L1_NONCE);
  }

  const [verifier, lineaRollupImplementation, proxyAdmin] = await Promise.all([
    deployContractFromArtifacts(verifierArtifacts.abi, verifierArtifacts.bytecode, wallet, {
      nonce: walletNonce,
      gasPrice,
    }),
    deployContractFromArtifacts(LineaRollupV6Abi, LineaRollupV6Bytecode, wallet, {
      nonce: walletNonce + 1,
      gasPrice,
    }),
    deployContractFromArtifacts(ProxyAdminAbi, ProxyAdminBytecode, wallet, { nonce: walletNonce + 2, gasPrice }),
  ]);

  const proxyAdminAddress = await proxyAdmin.getAddress();
  const verifierAddress = await verifier.getAddress();
  const lineaRollupImplementationAddress = await lineaRollupImplementation.getAddress();

  console.log(`${verifierName} deployed: address=${verifierAddress}`);
  console.log(`L1 ProxyAdmin deployed: address=${proxyAdminAddress}`);

  const initializer = getInitializerData(LineaRollupV6Abi, "initialize", [
    {
      initialStateRootHash: lineaRollupInitialStateRootHash,
      initialL2BlockNumber: lineaRollupInitialL2BlockNumber,
      genesisTimestamp: lineaRollupGenesisTimestamp,
      defaultVerifier: verifierAddress,
      rateLimitPeriodInSeconds: lineaRollupRateLimitPeriodInSeconds,
      rateLimitAmountInWei: lineaRollupRateLimitAmountInWei,
      roleAddresses,
      pauseTypeRoles,
      unpauseTypeRoles,
      fallbackOperator: multiCallAddress,
      defaultAdmin: lineaRollupSecurityCouncil,
    },
  ]);

  const proxyContract = await deployContractFromArtifacts(
    TransparentUpgradeableProxyAbi,
    TransparentUpgradeableProxyBytecode,
    wallet,
    lineaRollupImplementationAddress,
    proxyAdminAddress,
    initializer,
    { gasPrice },
  );

  const proxyContractAddress = await proxyContract.getAddress();
  const txReceipt = await proxyContract.deploymentTransaction()?.wait();

  if (!txReceipt) {
    throw "Contract deployment transaction receipt not found.";
  }

  console.log(`${lineaRollupName} deployed: address=${proxyContractAddress} blockNumber=${txReceipt.blockNumber}`);
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
