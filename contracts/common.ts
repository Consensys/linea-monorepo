import fs from "fs";

const MAX_GAS_LIMIT = process.env.TX_GAS_LIMIT ? parseInt(process.env.TX_GAS_LIMIT) : 500000000;

/**
 * Helper function to deal with account JSON files
 * @param {string} filePath full path to file with private key
 */
function getPrivateKeyFromFile(filePath: string): string {
  let privateKey;
  if (fs.existsSync(filePath)) {
    const account = fs.readFileSync(filePath, "utf8");
    privateKey = JSON.parse(account).account_key.priv_key;
  } else {
    console.warn(`Path ${filePath} with private key does not exist`);
  }
  return privateKey;
}

/**
 * Get private key of contract owner, based on several assumptions
 */
function getContractOwnerPrivateKey(
  keyFile: string,
  bridgeOwnerPathOverride: string,
  bridgeOwnerPrivateKeyOverride: string,
): string {
  let contractOwnerJsonPath;

  if (bridgeOwnerPrivateKeyOverride) {
    return bridgeOwnerPrivateKeyOverride;
  } else if (bridgeOwnerPathOverride && fs.existsSync(bridgeOwnerPathOverride)) {
    contractOwnerJsonPath = bridgeOwnerPathOverride;
  } else {
    contractOwnerJsonPath = keyFile;
  }
  return getPrivateKeyFromFile(contractOwnerJsonPath);
}

function getBlockchainNode(): string {
  return process.env.BLOCKCHAIN_NODE || "http://127.0.0.1:8545";
}

function getL2BlockchainNode(): string | undefined {
  return process.env.L2_BLOCKCHAIN_NODE;
}

/**
 * Returns path to JSON file with address of deployed rollup smart config, which contains address and ABI
 * @returns {string}
 */
function getRollupContractConfigPath(): string {
  return process.env.SMC_CONFIG_PATH || "/smart_contract/data/smc_config.json";
}

function getRollupJsonPath(): string {
  return process.env.ROLLUP_JSON_PATH || "/smart_contract/data/rollup.json";
}

export {
  MAX_GAS_LIMIT,
  getBlockchainNode,
  getContractOwnerPrivateKey,
  getL2BlockchainNode,
  getRollupContractConfigPath,
  getRollupJsonPath
};

