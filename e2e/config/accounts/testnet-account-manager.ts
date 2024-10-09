import { ethers, Provider, Wallet } from "ethers";
import Account from "./account";
import { AccountManager } from "./account-manager";

// Helper function to parse Ether amounts
function etherToWei(amount: number): bigint {
  return ethers.parseEther(amount.toString());
}

function getTransactionManager(provider: Provider, privateKey: string): Wallet {
  if (!privateKey.startsWith("0x")) {
    privateKey = "0x" + privateKey;
  }
  let keyWithoutPrefix = privateKey.slice(2);

  // Pad the private key to 64 hex characters (32 bytes) if it's shorter
  if (keyWithoutPrefix.length < 64) {
    keyWithoutPrefix = keyWithoutPrefix.padStart(64, "0");
  }
  return new Wallet(`0x${keyWithoutPrefix}`, provider);
}

class TestnetAccountManager implements AccountManager {
  private readonly chainId: number;
  private readonly whaleAccounts: Account[];
  private provider: Provider;
  private txManagers: Wallet[];

  constructor(provider: Provider, whaleAccounts: Account[]) {
    this.provider = provider;
    this.whaleAccounts = whaleAccounts;
    this.txManagers = this.whaleAccounts.map((account) => getTransactionManager(this.provider, account.privateKey));
  }

  selectWhaleAccount(): { account: Account; txManager: Wallet } {
    const testWorkerId = process.pid;
    const accountIndex = testWorkerId % this.whaleAccounts.length;
    const whaleAccount = this.whaleAccounts[accountIndex];
    const whaleTxManager = this.txManagers[accountIndex];

    return { account: whaleAccount, txManager: whaleTxManager };
  }

  whaleAccount(): Account {
    return this.selectWhaleAccount().account;
  }

  async generateAccount(initialBalanceWei = etherToWei(10)): Promise<Wallet> {
    const accounts = await this.generateAccounts(1, initialBalanceWei);
    return this.getTransactionManager(accounts[0]);
  }

  async generateAccounts(numberOfAccounts: number, initialBalanceWei = etherToWei(10)): Promise<Wallet[]> {
    const { account: whaleAccount, txManager: whaleTxManager } = this.selectWhaleAccount();

    console.log(
      `Generating accounts: chainId=${this.chainId} numberOfAccounts=${numberOfAccounts} whaleAccount=${whaleAccount.address}`,
    );

    const accounts: Account[] = [];

    for (let i = 0; i < numberOfAccounts; i++) {
      const randomBytes = ethers.randomBytes(32);
      const randomPrivKey = ethers.hexlify(randomBytes);
      const newAccount = new Account(randomPrivKey, ethers.computeAddress(randomPrivKey));
      accounts.push(newAccount);

      try {
        const tx = {
          to: newAccount.address,
          value: initialBalanceWei,
          gasPrice: ethers.parseUnits("300", "gwei"),
          gasLimit: 21000n,
        };
        const transactionResponse = await whaleTxManager.sendTransaction(tx);
        console.log(
          `Waiting for account funding: newAccount=${newAccount.address} txHash=${transactionResponse.hash} whaleAccount=${whaleAccount.address}`,
        );
        const receipt = await transactionResponse.wait();

        if (!receipt) {
          throw new Error(`Transaction failed to be mined`);
        }

        if (receipt.status !== 1) {
          throw new Error(`Transaction failed with status ${receipt.status}`);
        }
        console.log(
          `Account funded: newAccount=${newAccount.address} balance=${(
            await this.provider.getBalance(newAccount.address)
          ).toString()} wei`,
        );
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } catch (error: any) {
        const whaleBalance = await this.provider.getBalance(whaleAccount.address);
        throw new Error(
          `Failed to send funds from accAddress=${whaleAccount.address}, accBalance=${whaleBalance.toString()}, accPrivKey=0x...${whaleAccount.privateKey.slice(
            -8,
          )}, error: ${error.message}`,
        );
      }
      // finally {
      //   this.whaleAccountsInUse.delete(whaleAccount.address);
      // }
    }

    return accounts.map((account) => getTransactionManager(this.provider, account.privateKey));
  }

  getTransactionManager(account: Account): Wallet {
    return getTransactionManager(this.provider, account.privateKey);
  }
}

export { TestnetAccountManager };
