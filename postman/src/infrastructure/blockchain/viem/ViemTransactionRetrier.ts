import { type PublicClient, type WalletClient } from "viem";

import { ITransactionRetrier } from "../../../core/services/ITransactionRetrier";
import { Address, Hash, TransactionSubmission } from "../../../core/types";

const BASE_BUMP_PERCENT = 10n;

export class ViemTransactionRetrier implements ITransactionRetrier {
  constructor(
    private readonly publicClient: PublicClient,
    private readonly walletClient: WalletClient,
    private readonly signerAddress: Address,
    private readonly maxFeePerGasCap: bigint,
  ) {}

  public async retryWithHigherFee(transactionHash: Hash, attempt: number): Promise<TransactionSubmission> {
    const transaction = await this.publicClient.getTransaction({ hash: transactionHash });
    if (!transaction) {
      throw new Error(`Transaction with hash ${transactionHash} not found.`);
    }

    const { maxFeePerGas, maxPriorityFeePerGas } = await this.computeBumpedFees(
      transaction.maxFeePerGas,
      transaction.maxPriorityFeePerGas,
      attempt,
    );

    const txHash = await this.walletClient.sendTransaction({
      account: transaction.from,
      to: transaction.to ?? undefined,
      value: transaction.value,
      data: transaction.input,
      nonce: transaction.nonce,
      gas: transaction.gas,
      maxFeePerGas,
      maxPriorityFeePerGas,
      chain: null,
    });

    return {
      hash: txHash,
      nonce: transaction.nonce,
      gasLimit: transaction.gas,
      maxFeePerGas,
      maxPriorityFeePerGas,
    };
  }

  public async cancelTransaction(nonce: number): Promise<Hash> {
    const { maxFeePerGas, maxPriorityFeePerGas } = await this.getAggressiveFees();

    return this.walletClient.sendTransaction({
      account: this.signerAddress,
      to: this.signerAddress,
      value: 0n,
      data: "0x",
      nonce,
      gas: 21_000n,
      maxFeePerGas,
      maxPriorityFeePerGas,
      chain: null,
    });
  }

  private async computeBumpedFees(
    originalMaxFee: bigint | null | undefined,
    originalMaxPriority: bigint | null | undefined,
    attempt: number,
  ): Promise<{ maxFeePerGas: bigint; maxPriorityFeePerGas: bigint }> {
    if (!originalMaxFee || !originalMaxPriority) {
      return this.getAggressiveFees();
    }

    // Exponential bump: baseBump * 2^(attempt-1), minimum 10%
    const multiplier = BigInt(Math.max(1, 1 << (attempt - 1)));
    const bumpPercent = BASE_BUMP_PERCENT * multiplier;
    const bump = bumpPercent + 100n;

    let maxFeePerGas = (originalMaxFee * bump) / 100n;
    let maxPriorityFeePerGas = (originalMaxPriority * bump) / 100n;

    if (maxFeePerGas > this.maxFeePerGasCap) maxFeePerGas = this.maxFeePerGasCap;
    if (maxPriorityFeePerGas > this.maxFeePerGasCap) maxPriorityFeePerGas = this.maxFeePerGasCap;

    return { maxFeePerGas, maxPriorityFeePerGas };
  }

  private async getAggressiveFees(): Promise<{ maxFeePerGas: bigint; maxPriorityFeePerGas: bigint }> {
    const fees = await this.publicClient.estimateFeesPerGas();
    let maxFeePerGas = fees.maxFeePerGas * 2n;
    let maxPriorityFeePerGas = fees.maxPriorityFeePerGas * 2n;

    if (maxFeePerGas > this.maxFeePerGasCap) maxFeePerGas = this.maxFeePerGasCap;
    if (maxPriorityFeePerGas > this.maxFeePerGasCap) maxPriorityFeePerGas = this.maxFeePerGasCap;

    return { maxFeePerGas, maxPriorityFeePerGas };
  }
}
