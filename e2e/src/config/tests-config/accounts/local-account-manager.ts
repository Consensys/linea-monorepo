import { ethers, Provider } from "ethers";
import Account from "./account";
import { AccountManager } from "./account-manager";
import { readJsonFile } from "../../../common/utils";

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

function readGenesisFileAccounts(genesisJson: GenesisJson): Account[] {
  const alloc = genesisJson.alloc;
  const accounts: Account[] = [];
  for (const address in alloc) {
    const accountData = alloc[address];
    if (accountData.privateKey) {
      const addr = ethers.getAddress(address);
      accounts.push(new Account(accountData.privateKey, addr));
    }
  }
  return accounts;
}

class LocalAccountManager extends AccountManager {
  constructor(provider: Provider, genesisFilePath: string) {
    const genesisJson = readJsonFile(genesisFilePath);
    const genesis = genesisJson as GenesisJson;
    const chainId = genesis.config.chainId;
    const whaleAccounts = readGenesisFileAccounts(genesis);

    super(provider, whaleAccounts, chainId);
  }
}

export { LocalAccountManager };
