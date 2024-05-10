import { JsonRpcProvider } from "ethers";
import { FeeEstimationError } from "../../core/errors/GasFeeErrors";
import {
  DEFAULT_ENFORCE_MAX_GAS_FEE,
  DEFAULT_GAS_ESTIMATION_PERCENTILE,
  DEFAULT_MAX_FEE_PER_GAS,
} from "../../core/constants";
import { FeeHistory, Fees, IEIP1559GasProvider } from "../../core/clients/blockchain/IEIP1559GasProvider";

export class EIP1559GasProvider implements IEIP1559GasProvider {
  private feesCache: Fees;
  private cacheIsValidForBlockNumber: bigint;
  private gasEstimationPercentile: number;
  private isMaxGasFeeEnforced: boolean;
  protected readonly maxFeePerGas: bigint;

  constructor(
    protected readonly provider: JsonRpcProvider,
    maxFeePerGas?: bigint,
    gasEstimationPercentile?: number,
    enforceMaxGasFee?: boolean,
  ) {
    this.maxFeePerGas = BigInt(maxFeePerGas ?? DEFAULT_MAX_FEE_PER_GAS);
    this.gasEstimationPercentile = gasEstimationPercentile ?? DEFAULT_GAS_ESTIMATION_PERCENTILE;
    this.cacheIsValidForBlockNumber = 0n;
    this.feesCache = { maxFeePerGas: this.maxFeePerGas };
    this.isMaxGasFeeEnforced = enforceMaxGasFee ?? DEFAULT_ENFORCE_MAX_GAS_FEE;
  }

  /**
   * Fetches EIP-1559 gas fee estimates.
   *
   * This method uses the `eth_feeHistory` RPC endpoint to fetch historical gas fee data and calculates the
   * `maxPriorityFeePerGas` and `maxFeePerGas` based on the specified percentile. If `isMaxGasFeeEnforced` is true,
   * it returns the `maxFeePerGas` as configured in the constructor. Otherwise, it calculates the fees based on
   * the network conditions.
   *
   * The method caches the fee estimates and only fetches new data if the current block number has changed since
   * the last fetch. This reduces the number of RPC calls made to fetch fee data.
   *
   * @param {number} [percentile=this.gasEstimationPercentile] - The percentile value to sample from each block's effective priority fees.
   * @returns {Promise<Fees>} A promise that resolves to an object containing the `maxPriorityFeePerGas` and the `maxFeePerGas`.
   */
  public async get1559Fees(percentile: number = this.gasEstimationPercentile): Promise<Fees> {
    if (this.isMaxGasFeeEnforced) {
      return {
        maxPriorityFeePerGas: this.maxFeePerGas,
        maxFeePerGas: this.maxFeePerGas,
      };
    }

    const currentBlockNumber = await this.provider.getBlockNumber();
    if (this.cacheIsValidForBlockNumber < BigInt(currentBlockNumber)) {
      const { reward, baseFeePerGas }: FeeHistory = await this.provider.send("eth_feeHistory", [
        "0x4",
        "latest",
        [percentile],
      ]);

      const maxPriorityFeePerGas =
        reward.reduce((acc: bigint, currentValue: string[]) => acc + BigInt(currentValue[0]), 0n) /
        BigInt(reward.length);

      if (maxPriorityFeePerGas && maxPriorityFeePerGas > this.maxFeePerGas) {
        throw new FeeEstimationError(
          `Estimated miner tip of ${maxPriorityFeePerGas} exceeds configured max fee per gas of ${this.maxFeePerGas}!`,
        );
      }

      this.cacheIsValidForBlockNumber = BigInt(currentBlockNumber);

      const maxFeePerGas = BigInt(baseFeePerGas[baseFeePerGas.length - 1]) * 2n + maxPriorityFeePerGas;

      if (maxFeePerGas > 0n && maxPriorityFeePerGas > 0n) {
        this.feesCache = {
          maxPriorityFeePerGas,
          maxFeePerGas: maxFeePerGas > this.maxFeePerGas ? this.maxFeePerGas : maxFeePerGas,
        };
      } else {
        this.feesCache = {
          maxFeePerGas: this.maxFeePerGas,
        };
      }
    }
    return this.feesCache;
  }
}
