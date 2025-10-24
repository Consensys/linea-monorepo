import { Address, TransactionReceipt, BaseError, ContractFunctionRevertedError } from "viem";
import { ILidoAccountingReportClient } from "../core/clients/ILidoAccountingReportClient.js";
import { ILazyOracle, UpdateVaultDataParams } from "../core/clients/contracts/ILazyOracle.js";
import { getReportProofByVault } from "@lidofinance/lsv-cli/dist/utils/report/report-proof.js";
import { ILogger, IRetryService, bigintReplacer } from "@consensys/linea-shared-utils";

export class LidoAccountingReportClient implements ILidoAccountingReportClient {
  private latestSubmitVaultReportParams?: UpdateVaultDataParams;

  constructor(
    private readonly logger: ILogger,
    private readonly retryService: IRetryService,
    private readonly lazyOracleContractClient: ILazyOracle<TransactionReceipt>,
    private readonly ipfsGatewayUrl: string,
    private readonly vault: Address,
  ) {}

  async getLatestSubmitVaultReportParams(): Promise<UpdateVaultDataParams> {
    const latestReportData = await this.lazyOracleContractClient.latestReportData();
    const reportProof = await this.retryService.retry(() => getReportProofByVault({
      vault: this.vault,
      cid: latestReportData.reportCid,
      gateway: this.ipfsGatewayUrl,
    }));

    const params: UpdateVaultDataParams = {
      vault: this.vault,
      totalValue: BigInt(reportProof.data.totalValueWei),
      cumulativeLidoFees: BigInt(reportProof.data.fee),
      liabilityShares: BigInt(reportProof.data.liabilityShares),
      maxLiabilityShares: BigInt(reportProof.data.maxLiabilityShares),
      slashingReserve: BigInt(reportProof.data.slashingReserve),
      proof: reportProof.proof,
    };

    this.latestSubmitVaultReportParams = params;

    this.logger.info(`getLatestSubmitVaultReportParams set latestSubmitVaultReportParams=${JSON.stringify(params, bigintReplacer, 2)}`);
    return params;
  }

  // Return true is simulation succeeded, false otherwise
  async isSimulateSubmitLatestVaultReportSuccessful(): Promise<boolean> {
    const params = await this._getLatestSubmitVaultReportParams();

    try {
      await this.lazyOracleContractClient.simulateUpdateVaultData(params);
      this.logger.info("Successful isSimulateSubmitLatestVaultReportSuccessful");
      return true;
    } catch (err) {
      this.logger.error("Failed isSimulateSubmitLatestVaultReportSuccessful");
      if (err instanceof ContractFunctionRevertedError) {
        this.logger.error("❌ Reverted:", { shortMessage: err.shortMessage });
        this.logger.error("Reason:", { reason: err.data?.errorName || err.message });
      } else if (err instanceof BaseError) {
        this.logger.error("⚠️ Other error:", { shortMessage: err.shortMessage });
      } else {
        this.logger.error("Unexpected error:", { error: err });
      }
      return false;
    }
  }

  async submitLatestVaultReport(): Promise<void> {
    const params = await this._getLatestSubmitVaultReportParams();
    await this.lazyOracleContractClient.updateVaultData(params);
  }

  private async _getLatestSubmitVaultReportParams(): Promise<UpdateVaultDataParams> {
    if (!this.latestSubmitVaultReportParams) {
      this.latestSubmitVaultReportParams = await this.getLatestSubmitVaultReportParams();
    }

    return this.latestSubmitVaultReportParams;
  }
}
