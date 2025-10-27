import { IMetricsService } from "@consensys/linea-shared-utils";
import { BlockchainClientMetrics } from "../../metrics/BlockchainClientMetrics.js";
import { IBlockchainClientMetricsUpdater } from "../../metrics/IBlockchainClientMetricsUpdater.js";

export class BlockchainClientMetricsUpdater implements IBlockchainClientMetricsUpdater {
  constructor(private readonly metricsService: IMetricsService<BlockchainClientMetrics>) {
    this.metricsService.createCounter(
      BlockchainClientMetrics.TransactionFees,
      "Transaction fees paid (gwei) by automation per vault",
    );
  }

  public addTransactionFees(amountGwei: number): void {
    if (amountGwei <= 0) return;
    this.metricsService.incrementCounter(BlockchainClientMetrics.TransactionFees, {}, amountGwei);
  }
}
