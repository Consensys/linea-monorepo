import { Wallet } from "ethers";
import Account from "./account";

interface AccountManager {
  whaleAccount(): Account;
  generateAccount(initialBalanceWei?: bigint): Promise<Wallet>;
  generateAccounts(numberOfAccounts: number, initialBalanceWei?: bigint): Promise<Wallet[]>;
  getTransactionManager(account: Account): Wallet;
}

export type { AccountManager };
