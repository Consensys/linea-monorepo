import { ContractFactory, ethers } from "ethers";
import fs from "fs";
import path from "path";
import * as dotenv from "dotenv";
import { abi as ProxyAdminAbi, bytecode as ProxyAdminBytecode } from "./static-artifacts/ProxyAdmin.json";
import {
  abi as TransparentUpgradeableProxyAbi,
  bytecode as TransparentUpgradeableProxyBytecode,
} from "./static-artifacts/TransparentUpgradeableProxy.json";
import { getRequiredEnvVar } from "../common/helpers/environment";

dotenv.config();

function findContractArtifacts(
  folderPath: string,
  contractName: string,
): { abi: ethers.Interface | ethers.InterfaceAbi; bytecode: ethers.BytesLike } {
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
  const verifierName = getRequiredEnvVar("VERIFIER_CONTRACT_NAME");
  const LineaRollup_initialStateRootHash = getRequiredEnvVar("LINEA_ROLLUP_INITIAL_STATE_ROOT_HASH");
  const LineaRollup_initialL2BlockNumber = getRequiredEnvVar("LINEA_ROLLUP_INITIAL_L2_BLOCK_NUMBER");
  const LineaRollup_securityCouncil = getRequiredEnvVar("LINEA_ROLLUP_SECURITY_COUNCIL");
  const LineaRollup_operators = getRequiredEnvVar("LINEA_ROLLUP_OPERATORS");
  const LineaRollup_rateLimitPeriodInSeconds = getRequiredEnvVar("LINEA_ROLLUP_RATE_LIMIT_PERIOD");
  const LineaRollup_rateLimitAmountInWei = getRequiredEnvVar("LINEA_ROLLUP_RATE_LIMIT_AMOUNT");
  const LineaRollup_genesisTimestamp = getRequiredEnvVar("LINEA_ROLLUP_GENESIS_TIMESTAMP");
  const lineaRollupName = "LineaRollupV5";

  const verifierArtifacts = findContractArtifacts(path.join(__dirname, "./shared-artifacts"), verifierName);
  const lineaRollupArtifacts = findContractArtifacts(path.join(__dirname, "./shared-artifacts"), lineaRollupName);

  const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);
  const wallet = new ethers.Wallet(process.env.PRIVATE_KEY!, provider);

  const walletNonce = await wallet.getNonce();

  const [verifierAddress, lineaRollupImplAddress, proxyAdminAddress] = await Promise.all([
    deployContract(verifierArtifacts.abi, verifierArtifacts.bytecode, wallet, { nonce: walletNonce }),
    deployContract(lineaRollupArtifacts.abi, lineaRollupArtifacts.bytecode, wallet, { nonce: walletNonce + 1 }),
    deployContract(ProxyAdminAbi, ProxyAdminBytecode, wallet, { nonce: walletNonce + 2 }),
  ]);

  console.log(`${verifierName} contract deployed at ${verifierAddress}`);
  console.log(`${lineaRollupName} Implementation contract deployed at ${lineaRollupImplAddress}`);
  console.log(`Proxy admin contract deployed at ${proxyAdminAddress}`);

  const initializer = new ContractFactory(
    lineaRollupArtifacts.abi,
    lineaRollupArtifacts.bytecode,
  ).interface.encodeFunctionData("initialize", [
    LineaRollup_initialStateRootHash,
    LineaRollup_initialL2BlockNumber,
    verifierAddress,
    LineaRollup_securityCouncil,
    LineaRollup_operators?.split(","),
    LineaRollup_rateLimitPeriodInSeconds,
    LineaRollup_rateLimitAmountInWei,
    LineaRollup_genesisTimestamp,
  ]);

  const proxyAddress = await deployContract(
    TransparentUpgradeableProxyAbi,
    TransparentUpgradeableProxyBytecode,
    wallet,
    lineaRollupImplAddress,
    proxyAdminAddress,
    initializer,
  );

  console.log(`${lineaRollupName} Proxy contract deployed at ${proxyAddress}`);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
