import { readFileSync } from "fs";
import Account from "./account";
import { AccountManager } from "./account-manager";
import { Client, getAddress } from "viem";

interface GenesisJson {
  config: {
    chainId: number;
  };
  alloc: {
    [address: string]: {
      privateKey?: string;
    };
  };
}

function readJsonFile(filePath: string): unknown {
  const data = readFileSync(filePath, "utf8");
  return JSON.parse(data);
}

function readGenesisFileAccounts(genesisJson: GenesisJson): Account[] {
  const alloc = genesisJson.alloc;
  const accounts: Account[] = [];
  for (const address in alloc) {
    const accountData = alloc[address];
    if (accountData.privateKey) {
      const prefixedAddress = address.startsWith("0x") ? address : `0x${address}`;
      const addr = getAddress(prefixedAddress);
      accounts.push(new Account(accountData.privateKey as `0x${string}`, addr));
    }
  }
  return accounts;
}

class GenesisBasedAccountManager extends AccountManager {
  constructor(client: Client, genesisFilePath: string) {
    const genesisJson = readJsonFile(genesisFilePath);
    const genesis = genesisJson as GenesisJson;
    const chainId = genesis.config.chainId;
    const whaleAccounts = readGenesisFileAccounts(genesis);

    super(client, whaleAccounts, chainId);
  }
}

export { GenesisBasedAccountManager };
