import { Provider } from "ethers";
import Account from "./account";
import { AccountManager } from "./account-manager";

class TestnetAccountManager extends AccountManager {
  constructor(provider: Provider, whaleAccounts: Account[], chainId: number) {
    super(provider, whaleAccounts, chainId);
  }
}

export { TestnetAccountManager };
