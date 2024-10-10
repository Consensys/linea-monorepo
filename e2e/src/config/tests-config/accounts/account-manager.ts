import { ethers, Provider, Wallet } from "ethers";
import Account from "./account";
import { etherToWei } from "../../../common/utils";

interface IAccountManager {
  whaleAccount(accIndex?: number): Wallet;
  generateAccount(initialBalanceWei?: bigint): Promise<Wallet>;
  generateAccounts(numberOfAccounts: number, initialBalanceWei?: bigint): Promise<Wallet[]>;
  getWallet(account: Account): Wallet;
}

function getWallet(provider: Provider, privateKey: string): Wallet {
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

abstract class AccountManager implements IAccountManager {
  protected readonly chainId: number;
  protected readonly whaleAccounts: Account[];
  protected provider: Provider;
  protected accountWallets: Wallet[];

  constructor(provider: Provider, whaleAccounts: Account[], chainId: number) {
    this.provider = provider;
    this.whaleAccounts = whaleAccounts;
    this.chainId = chainId;
    this.accountWallets = this.whaleAccounts.map((account) => getWallet(this.provider, account.privateKey));
  }

  selectWhaleAccount(accIndex?: number): { account: Account; accountWallet: Wallet } {
    if (accIndex) {
      return { account: this.whaleAccounts[accIndex], accountWallet: this.accountWallets[accIndex] };
    }
    const workerIdEnv = process.env.JEST_WORKER_ID || "1";
    const workerId = parseInt(workerIdEnv, 10) - 1;

    const accountIndex = workerId;
    const whaleAccount = this.whaleAccounts[accountIndex];
    const whaleTxManager = this.accountWallets[this.whaleAccounts.indexOf(whaleAccount)];
    return { account: whaleAccount, accountWallet: whaleTxManager };
  }

  whaleAccount(accIndex?: number): Wallet {
    return this.selectWhaleAccount(accIndex).accountWallet;
  }

  async generateAccount(initialBalanceWei = etherToWei("10")): Promise<Wallet> {
    const accounts = await this.generateAccounts(1, initialBalanceWei);
    return this.getWallet(accounts[0]);
  }

  async generateAccounts(
    numberOfAccounts: number,
    initialBalanceWei = etherToWei("10"),
    retryDelayMs = 3_000,
  ): Promise<Wallet[]> {
    const { account: whaleAccount, accountWallet: whaleAccountWallet } = this.selectWhaleAccount();

    console.log(
      `Generating accounts: chainId=${this.chainId} numberOfAccounts=${numberOfAccounts} whaleAccount=${whaleAccount.address}`,
    );

    const accounts: Account[] = [];

    for (let i = 0; i < numberOfAccounts; i++) {
      const randomBytes = ethers.randomBytes(32);
      const randomPrivKey = ethers.hexlify(randomBytes);
      const newAccount = new Account(randomPrivKey, ethers.computeAddress(randomPrivKey));
      accounts.push(newAccount);

      let success = false;
      while (!success) {
        try {
          const tx = {
            to: newAccount.address,
            value: initialBalanceWei,
            gasPrice: ethers.parseUnits("300", "gwei"),
            gasLimit: 21000n,
          };
          const transactionResponse = await whaleAccountWallet.sendTransaction(tx);
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
          success = true;
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } catch (error: any) {
          console.log(
            `Failed to send funds from accAddress=${whaleAccount.address}. Retrying funding in ${retryDelayMs}ms...`,
          );
          await new Promise((resolve) => setTimeout(resolve, retryDelayMs));
        }
      }
    }
    return accounts.map((account) => getWallet(this.provider, account.privateKey));
  }

  getWallet(account: Account): Wallet {
    return getWallet(this.provider, account.privateKey);
  }
}

export { AccountManager };
