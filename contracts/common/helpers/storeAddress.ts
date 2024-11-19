/* eslint-disable @typescript-eslint/no-var-requires */
import fs from "fs";
import { ethers } from "hardhat";
import { DeploymentsExtension } from "hardhat-deploy/types";
import path from "path";
const editJsonFile = require("edit-json-file");

export const tryStoreAddress = async (
  networkName: string,
  contractName: string,
  address: string,
  transactionHash: string,
) => {
  if (process.env.SAVE_ADDRESS) {
    const network = await ethers.provider.getNetwork();

    const ContractFactory = await ethers.getContractFactory(contractName);
    const dirPath = path.join(__dirname, "..", "..", "deployments", `${networkName}`);

    if (!fs.existsSync(dirPath)) {
      fs.mkdirSync(dirPath, { recursive: true });
    }

    const chainIdFile = path.join(dirPath, ".chainId");
    // if .chainId does not exist, add it
    if (!fs.existsSync(chainIdFile)) {
      fs.writeFileSync(chainIdFile, network.chainId.toString());
    }

    fs.writeFileSync(
      path.join(dirPath, `${contractName}.json`),
      JSON.stringify(
        {
          address: address,
          abi: ContractFactory.interface.formatJson(),
          transactionHash: transactionHash,
        },
        null,
        2,
      ),
    );
  }
};

export const tryStoreProxyAdminAddress = async (networkName: string, contractName: string, address: string) => {
  if (process.env.SAVE_ADDRESS) {
    const network = await ethers.provider.getNetwork();

    const dirPath = path.join(__dirname, "..", "deployments", `${networkName}`);

    if (!fs.existsSync(dirPath)) {
      fs.mkdirSync(dirPath, { recursive: true });
    }

    const chainIdFile = path.join(dirPath, ".chainId");
    // if .chainId does not exist, add it
    if (!fs.existsSync(chainIdFile)) {
      fs.writeFileSync(chainIdFile, network.chainId.toString());
    }

    fs.writeFileSync(
      path.join(dirPath, `${contractName}ProxyAdmin.json`),
      JSON.stringify(
        {
          address: address,
          abi: null,
          transactionHash: null,
        },
        null,
        2,
      ),
    );
  }
};

export const storeConstructorArgs = async (contractName: string, args: unknown[]): Promise<string> => {
  const filename = `./${contractName}ConstructorArgs.js`;
  console.log(`Generating constructor arguments file ${filename}`);

  const file = editJsonFile(filename);
  file.write(`module.exports = ${JSON.stringify(args)}`);

  return filename;
};

export const deleteConstructorArgs = (filePath: string) => {
  console.log(`Deleting constructor arguments file ${filePath}`);
  fs.unlink(filePath, function (err) {
    if (err) throw err;
    // if no error, file has been deleted successfully
    console.log("File deleted!");
  });
};

export const getDeployedContractAddress = async (
  contractName: string,
  deployments: DeploymentsExtension,
): Promise<string | undefined> => {
  const { get } = deployments;
  try {
    const deploymentDetails = await get(contractName);
    console.log(`Existing ${contractName} contract found.`, deploymentDetails.address);
    return deploymentDetails.address;
  } catch {
    // log error - existing contract not found
    console.log("Existing deployment not found for ", contractName);
  }
  return undefined;
};
