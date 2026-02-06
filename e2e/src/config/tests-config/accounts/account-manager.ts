import { Client, createNonceManager, Hex, PrivateKeyAccount, SendTransactionReturnType } from "viem";
import { privateKeyToAccount, generatePrivateKey, privateKeyToAddress } from "viem/accounts";
import { waitForTransactionReceipt } from "viem/actions";
import { jsonRpc } from "viem/nonce";

import Account from "./account";
import { AccountFundingService } from "./account-funding-service";
import { etherToWei, normalizeAddress } from "../../../common/utils";
import { createTestLogger } from "../../../config/logger";

import type { Logger } from "winston";

interface IAccountManager {
  whaleAccount(accIndex?: number): PrivateKeyAccount;
  generateAccount(initialBalanceWei?: bigint, accIndex?: number): Promise<PrivateKeyAccount>;
  generateAccounts(
    numberOfAccounts: number,
    initialBalanceWei?: bigint,
    accIndex?: number,
  ): Promise<PrivateKeyAccount[]>;
}

function formatPrivateKey(privateKey: string): Hex {
  if (!privateKey.startsWith("0x")) {
    privateKey = "0x" + privateKey;
  }
  let keyWithoutPrefix = privateKey.slice(2);

  // Pad the private key to 64 hex characters (32 bytes) if it's shorter
  if (keyWithoutPrefix.length < 64) {
    keyWithoutPrefix = keyWithoutPrefix.padStart(64, "0");
  }
  return `0x${keyWithoutPrefix}`;
}

abstract class AccountManager implements IAccountManager {
  protected readonly chainId: number;
  protected readonly whaleAccounts: Account[];
  protected readonly reservedAddresses: string[];
  protected client: Client;
  protected accountWallets: PrivateKeyAccount[];
  private logger: Logger;
  private fundingService: AccountFundingService;

  private readonly MAX_RETRIES = 5;
  private readonly RETRY_DELAY_MS = 1_000;

  constructor(client: Client, whaleAccounts: Account[], chainId: number, reservedAddresses: string[] = []) {
    this.client = client;
    this.whaleAccounts = whaleAccounts;
    this.chainId = chainId;
    this.reservedAddresses = reservedAddresses.map(normalizeAddress);
    this.accountWallets = this.whaleAccounts.map((account) => {
      const nonceManager = createNonceManager({
        source: jsonRpc(),
      });

      return privateKeyToAccount(formatPrivateKey(account.privateKey), { nonceManager });
    });

    this.logger = createTestLogger();
    this.fundingService = new AccountFundingService(this.client, this.chainId, this.logger, {
      retries: this.MAX_RETRIES,
      delayMs: this.RETRY_DELAY_MS,
    });
  }

  selectWhaleAccount(accIndex?: number): { account: Account; accountWallet: PrivateKeyAccount } {
    if (accIndex !== undefined) {
      return { account: this.whaleAccounts[accIndex], accountWallet: this.accountWallets[accIndex] };
    }

    const workerIdEnv = process.env.JEST_WORKER_ID ?? "1";
    const workerId = Number(workerIdEnv) - 1;

    const availableWhaleAccounts = this.whaleAccounts.filter(
      (account) => !this.reservedAddresses.includes(normalizeAddress(account.address)),
    );

    if (availableWhaleAccounts.length === 0) {
      throw new Error("No available whale accounts found after filtering reserved addresses.");
    }

    const isValidWorkerId = Number.isFinite(workerId) && workerId >= 0;
    if (!isValidWorkerId) {
      this.logger.warn(`Invalid JEST_WORKER_ID value. value=${workerIdEnv}`);
    }
    const accountIndex = isValidWorkerId ? workerId % availableWhaleAccounts.length : 0;
    const whaleAccount = availableWhaleAccounts[accountIndex];
    const whaleTxManager = this.accountWallets[this.whaleAccounts.indexOf(whaleAccount)];
    return { account: whaleAccount, accountWallet: whaleTxManager };
  }

  whaleAccount(accIndex?: number): PrivateKeyAccount {
    return this.selectWhaleAccount(accIndex).accountWallet;
  }

  async generateAccount(initialBalanceWei = etherToWei("10"), accIndex?: number): Promise<PrivateKeyAccount> {
    const accounts = await this.generateAccounts(1, initialBalanceWei, accIndex);
    return accounts[0];
  }

  async generateAccounts(
    numberOfAccounts: number,
    initialBalanceWei = etherToWei("10"),
    accIndex?: number,
  ): Promise<PrivateKeyAccount[]> {
    const { account: whaleAccount, accountWallet: whaleAccountWallet } = this.selectWhaleAccount(accIndex);

    this.logger.debug(
      `Generating accounts... chainId=${this.chainId} numberOfAccounts=${numberOfAccounts} whaleAccount=${whaleAccount.address}`,
    );

    const accountTransactionPairs: Array<{ account: Account; txPromise: Promise<SendTransactionReturnType | null> }> =
      [];

    for (let i = 0; i < numberOfAccounts; i++) {
      const randomPrivKey = generatePrivateKey();
      const newAccount = new Account(randomPrivKey, privateKeyToAddress(randomPrivKey));

      const txPromise = this.fundingService.fundAccount(
        whaleAccountWallet,
        whaleAccount.address,
        newAccount.address,
        initialBalanceWei,
      );

      accountTransactionPairs.push({ account: newAccount, txPromise });
    }

    const transactionResults = await Promise.all(accountTransactionPairs.map((pair) => pair.txPromise));

    const successfulTransactions = transactionResults.filter(
      (txResponse): txResponse is SendTransactionReturnType => txResponse !== null,
    );

    const failedCount = numberOfAccounts - successfulTransactions.length;
    if (failedCount > 0) {
      this.logger.warn(
        `Some accounts were not funded successfully. successful=${successfulTransactions.length} failed=${failedCount} expected=${numberOfAccounts}`,
      );
    }

    if (successfulTransactions.length === 0) {
      throw new Error(`Failed to fund any accounts. All ${numberOfAccounts} account funding attempts failed.`);
    }

    // Wait for successful transactions to be confirmed
    await Promise.all(successfulTransactions.map((tx) => waitForTransactionReceipt(this.client, { hash: tx })));

    // Return all accounts (both funded and unfunded) to maintain backward compatibility
    // Unfunded accounts will fail later when used, providing clearer error messages
    const allAccounts = accountTransactionPairs.map((pair) => pair.account);

    this.logger.debug(
      `${successfulTransactions.length}/${numberOfAccounts} accounts funded successfully. newAccounts=${allAccounts
        .map((account) => account.address)
        .join(", ")} balance=${initialBalanceWei.toString()} Wei`,
    );

    return allAccounts.map((account) => privateKeyToAccount(account.privateKey));
  }
}

export { AccountManager };
