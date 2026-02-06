import { Mutex } from "async-mutex";
import type { Logger } from "winston";
import { estimateLineaGas } from "../../../common/utils";
import { Address, Client, parseGwei, PrivateKeyAccount, SendTransactionReturnType } from "viem";
import { estimateFeesPerGas, sendTransaction } from "viem/actions";
import { RetryPolicy } from "./retry-policy";

type FeeData = {
  maxPriorityFeePerGas: bigint;
  maxFeePerGas: bigint;
};

export class AccountFundingService {
  private readonly whaleAccountMutex = new Mutex();
  private readonly retryPolicy: RetryPolicy;

  constructor(
    private readonly client: Client,
    private readonly chainId: number,
    private readonly logger: Logger,
    retryOptions: { retries: number; delayMs: number },
  ) {
    this.retryPolicy = new RetryPolicy(logger, retryOptions);
  }

  async fundAccount(
    whaleAccountWallet: PrivateKeyAccount,
    whaleAccountAddress: Address,
    targetAddress: Address,
    initialBalanceWei: bigint,
  ): Promise<SendTransactionReturnType | null> {
    const feeData = await this.estimateFees(whaleAccountWallet.address, targetAddress, initialBalanceWei);

    const sendWithRetry = async (): Promise<SendTransactionReturnType> => {
      return this.retryPolicy.execute(async () => {
        const release = await this.whaleAccountMutex.acquire();
        try {
          const transactionHash = await sendTransaction(this.client, {
            account: whaleAccountWallet,
            chain: this.client.chain,
            type: "eip1559",
            to: targetAddress,
            value: initialBalanceWei,
            maxPriorityFeePerGas: feeData.maxPriorityFeePerGas,
            maxFeePerGas: feeData.maxFeePerGas,
            gas: 21000n,
          });
          this.logger.debug(
            `Transaction sent. newAccount=${targetAddress} txHash=${transactionHash} whaleAccount=${whaleAccountAddress}`,
          );
          return transactionHash;
        } catch (error) {
          this.logger.warn(`sendTransaction failed for account=${targetAddress}. Error: ${(error as Error).message}`);
          throw error;
        } finally {
          release();
        }
      });
    };

    return sendWithRetry().catch((error) => {
      this.logger.error(
        `Failed to fund account after retries. address=${targetAddress} error=${(error as Error).message}`,
      );
      whaleAccountWallet.nonceManager?.reset({
        address: whaleAccountWallet.address,
        chainId: this.chainId,
      });
      return null;
    });
  }

  private async estimateFees(fromAddress: Address, toAddress: Address, value: bigint): Promise<FeeData> {
    if (this.chainId === 1337) {
      const feeData = await estimateLineaGas(this.client, {
        account: fromAddress,
        to: toAddress,
        value,
      });
      return {
        maxPriorityFeePerGas: feeData.maxPriorityFeePerGas,
        maxFeePerGas: feeData.maxFeePerGas,
      };
    }

    const feeData = await estimateFeesPerGas(this.client);
    return {
      maxPriorityFeePerGas: feeData.maxPriorityFeePerGas ?? parseGwei("1"),
      maxFeePerGas: feeData.maxFeePerGas ?? parseGwei("10"),
    };
  }
}
