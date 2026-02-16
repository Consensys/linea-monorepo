import { MINIMUM_MARGIN, PROFIT_MARGIN_MULTIPLIER } from "../constants";
import { BlockExtraData } from "../types";

export type ProfitabilityConfig = {
  profitMargin: number;
  maxClaimGasLimit: bigint;
  isPostmanSponsorshipEnabled: boolean;
  maxPostmanSponsorGasLimit: bigint;
};

export type L1ProfitabilityInput = {
  gasLimit: bigint;
  messageFee: bigint;
  maxFeePerGas: bigint;
};

export type L2ProfitabilityInput = {
  gasLimit: bigint;
  messageFee: bigint;
  compressedTransactionSize: number;
  blockExtraData: BlockExtraData;
};

export class ProfitabilityCalculator {
  constructor(private readonly config: ProfitabilityConfig) {}

  /**
   * Check if a message has zero fee and profit margin is configured (nonzero).
   * Returns true if the message fee is 0 and profitMargin is not disabled.
   */
  public hasZeroFee(messageFee: bigint): boolean {
    return messageFee === 0n && this.config.profitMargin !== 0;
  }

  /**
   * Calculate the gas estimation threshold (fee per gas unit).
   */
  public calculateGasEstimationThreshold(messageFee: bigint, gasLimit: bigint): number {
    return parseFloat((messageFee / gasLimit).toString());
  }

  /**
   * Determine if the gas limit exceeds the maximum allowed.
   * Returns the gas limit if within bounds, null otherwise.
   */
  public getGasLimit(gasLimit: bigint): bigint | null {
    return gasLimit <= this.config.maxClaimGasLimit ? gasLimit : null;
  }

  /**
   * Determine if a message qualifies for sponsorship.
   */
  public isForSponsorship(gasLimit: bigint, hasZeroFee: boolean, isUnderPriced: boolean): boolean {
    if (!this.config.isPostmanSponsorshipEnabled) return false;
    if (gasLimit > this.config.maxPostmanSponsorGasLimit) return false;
    if (hasZeroFee) return true;
    if (isUnderPriced) return true;
    return false;
  }

  /**
   * L1 (Ethereum) underpricing check.
   * The message is underpriced if the actual claiming cost exceeds the message fee.
   */
  public isL1UnderPriced(input: L1ProfitabilityInput): boolean {
    const { gasLimit, messageFee, maxFeePerGas } = input;
    const actualCost =
      gasLimit * maxFeePerGas * BigInt(Math.floor(this.config.profitMargin * PROFIT_MARGIN_MULTIPLIER));
    const maxFee = messageFee * BigInt(PROFIT_MARGIN_MULTIPLIER);
    return actualCost > maxFee;
  }

  /**
   * L2 (Linea) underpricing check.
   * Uses block extra data (variable/fixed cost) and compressed transaction size
   * to compute a priority fee, then compare against the message fee.
   */
  public isL2UnderPriced(input: L2ProfitabilityInput): boolean {
    const { gasLimit, messageFee, compressedTransactionSize, blockExtraData } = input;

    const priorityFee =
      (BigInt(MINIMUM_MARGIN * 10) *
        ((BigInt(blockExtraData.variableCost) * BigInt(compressedTransactionSize)) / gasLimit +
          BigInt(blockExtraData.fixedCost))) /
      10n;

    const actualCost = priorityFee * gasLimit * BigInt(Math.floor(this.config.profitMargin * PROFIT_MARGIN_MULTIPLIER));
    const maxFee = messageFee * BigInt(PROFIT_MARGIN_MULTIPLIER);
    return maxFee < actualCost;
  }
}
