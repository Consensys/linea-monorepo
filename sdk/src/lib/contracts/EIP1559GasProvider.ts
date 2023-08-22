import { JsonRpcProvider } from "@ethersproject/providers";
import { BigNumber } from "ethers";
import { FeeEstimationError } from "../utils/errors";
import { DEFAULT_GAS_ESTIMATION_PERCENTILE, DEFAULT_MAX_FEE_PER_GAS } from "../utils/constants";

type Fees = {
  maxFeePerGas: BigNumber;
  maxPriorityFeePerGas?: BigNumber;
};

type FeeHistory = {
  oldestBlock: number;
  reward: string[][];
  baseFeePerGas: string[];
  gasUsedRatio: number[];
};

export class EIP1559GasProvider {
  private feesCacheOld: Fees;
  private feesCache: Fees;
  private cacheIsValidForBlockNumber: BigNumber;
  private maxFeePerGasFromConfig: BigNumber;
  private gasEstimationPercentile: number;

  constructor(protected readonly provider: JsonRpcProvider, maxFeePerGas?: number, gasEstimationPercentile?: number) {
    this.maxFeePerGasFromConfig = BigNumber.from(maxFeePerGas ?? DEFAULT_MAX_FEE_PER_GAS);
    this.gasEstimationPercentile = gasEstimationPercentile ?? DEFAULT_GAS_ESTIMATION_PERCENTILE;
    this.cacheIsValidForBlockNumber = BigNumber.from(0);
    this.feesCacheOld = { maxFeePerGas: this.maxFeePerGasFromConfig };
    this.feesCache = { maxFeePerGas: this.maxFeePerGasFromConfig };
  }

  public async get1559Fees(percentile = this.gasEstimationPercentile): Promise<Fees> {
    const currentBlockNumber = await this.provider.getBlockNumber();
    if (this.cacheIsValidForBlockNumber.lt(currentBlockNumber)) {
      const { reward, baseFeePerGas }: FeeHistory = await this.provider.send("eth_feeHistory", [
        "0x4",
        "latest",
        [percentile],
      ]);

      const maxPriorityFeePerGas = reward
        .reduce((acc: BigNumber, currentValue: string[]) => acc.add(currentValue[0]), BigNumber.from(0))
        .div(reward.length);

      if (maxPriorityFeePerGas && maxPriorityFeePerGas.gt(this.maxFeePerGasFromConfig)) {
        throw new FeeEstimationError(
          `Estimated miner tip of ${maxPriorityFeePerGas} exceeds configured max fee per gas of ${this.maxFeePerGasFromConfig}!`,
        );
      }

      this.cacheIsValidForBlockNumber = BigNumber.from(currentBlockNumber);

      const maxFeePerGas = BigNumber.from(baseFeePerGas[baseFeePerGas.length - 1])
        .mul(2)
        .add(maxPriorityFeePerGas);

      if (maxFeePerGas.gt(0) && maxPriorityFeePerGas.gt(0)) {
        this.feesCache = {
          maxPriorityFeePerGas,
          maxFeePerGas: maxFeePerGas.gt(this.maxFeePerGasFromConfig) ? this.maxFeePerGasFromConfig : maxFeePerGas,
        };
      } else {
        this.feesCache = {
          maxFeePerGas: this.maxFeePerGasFromConfig,
        };
      }
    }
    return this.feesCache;
  }

  public async get1559FeesOld(percentile = this.gasEstimationPercentile): Promise<Fees> {
    const currentBlockNumber = await this.provider.getBlockNumber();
    if (this.cacheIsValidForBlockNumber.lt(currentBlockNumber)) {
      const feeHistory: FeeHistory = await this.provider.send("eth_feeHistory", ["0xa", "pending", [percentile]]);

      const maxPriorityFeePerGas = this.calculatePriorityFee(feeHistory);

      const maxFeePerGas = BigNumber.from(feeHistory.baseFeePerGas[feeHistory.baseFeePerGas.length - 1])
        .mul(2)
        .add(maxPriorityFeePerGas);

      if (maxPriorityFeePerGas.gt(this.maxFeePerGasFromConfig)) {
        throw new FeeEstimationError(
          `Estimated miner tip of ${maxPriorityFeePerGas} exceeds configured max fee per gas of ${this.maxFeePerGasFromConfig}!`,
        );
      }

      this.cacheIsValidForBlockNumber = BigNumber.from(currentBlockNumber);

      if (maxFeePerGas.gt(0) && maxPriorityFeePerGas.gt(0)) {
        this.feesCacheOld = {
          maxFeePerGas: maxFeePerGas.gt(this.maxFeePerGasFromConfig) ? this.maxFeePerGasFromConfig : maxFeePerGas,
          maxPriorityFeePerGas: maxPriorityFeePerGas,
        };
      } else {
        this.feesCacheOld = {
          maxFeePerGas: this.maxFeePerGasFromConfig,
        };
      }
    }
    return this.feesCacheOld;
  }

  private calculatePriorityFee(feeHistory: FeeHistory): BigNumber {
    const priorityFeeWeightedSum = feeHistory.reward.reduce(
      (a, v, index) => a + Number(v[0]) * feeHistory.gasUsedRatio[index],
      0,
    );
    const sumOfRatios = feeHistory.gasUsedRatio.reduce((a, v) => a + v, 0);

    if (sumOfRatios === 0) {
      return BigNumber.from(0);
    }

    return BigNumber.from(Math.round(priorityFeeWeightedSum / (sumOfRatios * feeHistory.gasUsedRatio.length)));
  }
}
