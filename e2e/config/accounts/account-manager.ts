import { Wallet } from "ethers";
import Account from "./account";

interface AccountManager {
  whaleAccount(accIndex?: number): Wallet;
  generateAccount(initialBalanceWei?: bigint): Promise<Wallet>;
  generateAccounts(numberOfAccounts: number, initialBalanceWei?: bigint): Promise<Wallet[]>;
  getTransactionManager(account: Account): Wallet;
}

export type { AccountManager };
