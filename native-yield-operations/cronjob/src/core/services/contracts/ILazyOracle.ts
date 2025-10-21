import { Address, Hex } from "viem";
import { LazyOracleReportData } from "../../entities";
export interface ILazyOracle<TransactionReceipt> {
  updateVaultData(
    vault: Address,
    totalValue: bigint,
    cumulativeLidoFees: bigint,
    liabilityShares: bigint,
    maxLiabilityShares: bigint,
    slashingReserve: bigint,
    proof: Hex[],
  ): Promise<TransactionReceipt | null>;
  latestReportData(): Promise<LazyOracleReportData>;
}

// retryTransactionWithHigherFee(transactionHash: string, priceBumpPercent?: number): Promise<TransactionResponse>;
// parseTransactionError(transactionHash: string): Promise<ErrorDescription | string>;
