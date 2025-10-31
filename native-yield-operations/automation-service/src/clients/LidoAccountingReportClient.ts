import { Address, TransactionReceipt, BaseError, ContractFunctionRevertedError } from "viem";
import { ILidoAccountingReportClient } from "../core/clients/ILidoAccountingReportClient.js";
import { ILazyOracle, UpdateVaultDataParams } from "../core/clients/contracts/ILazyOracle.js";
import { getReportProofByVault } from "@lidofinance/lsv-cli/dist/utils/report/report-proof.js";
import { ILogger, IRetryService, bigintReplacer } from "@consensys/linea-shared-utils";

export class LidoAccountingReportClient implements ILidoAccountingReportClient {
  private vaultReportByAddress = new Map<Address, UpdateVaultDataParams>();

  constructor(
    private readonly logger: ILogger,
    private readonly retryService: IRetryService,
    private readonly lazyOracleContractClient: ILazyOracle<TransactionReceipt>,
    private readonly ipfsGatewayUrl: string,
  ) {}

  async getLatestSubmitVaultReportParams(vault: Address): Promise<UpdateVaultDataParams> {
    const latestReportData = await this.lazyOracleContractClient.latestReportData();
    const reportProof = await this.retryService.retry(() =>
      getReportProofByVault({
        vault,
        cid: latestReportData.reportCid,
        gateway: this.ipfsGatewayUrl,
      }),
    );

    const params: UpdateVaultDataParams = {
      vault,
      totalValue: BigInt(reportProof.data.totalValueWei),
      cumulativeLidoFees: BigInt(reportProof.data.fee),
      liabilityShares: BigInt(reportProof.data.liabilityShares),
      maxLiabilityShares: BigInt(reportProof.data.maxLiabilityShares),
      slashingReserve: BigInt(reportProof.data.slashingReserve),
      proof: reportProof.proof,
    };

    this.vaultReportByAddress.set(vault, params);

    this.logger.info(
      `getLatestSubmitVaultReportParams for vault=${vault} latestSubmitVaultReportParams=${JSON.stringify(params, bigintReplacer, 2)}`,
    );
    return params;
  }

  // Uses latest known result of this.getLatestSubmitVaultReportParams()
  // Return true is simulation succeeded, false otherwise
  async isSimulateSubmitLatestVaultReportSuccessful(vault: Address): Promise<boolean> {
    const params = await this._getLatestSubmitVaultReportParams(vault);

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
        const decodedError = err.walk();
        this.logger.error("⚠️ Error:", { decodedError });
      } else {
        this.logger.error("Unexpected error:", { error: err });
      }
      return false;
    }
  }

  // Uses latest known result of this.getLatestSubmitVaultReportParams()
  async submitLatestVaultReport(vault: Address): Promise<void> {
    const params = await this._getLatestSubmitVaultReportParams(vault);
    await this.lazyOracleContractClient.updateVaultData(params);
  }

  private async _getLatestSubmitVaultReportParams(vault: Address): Promise<UpdateVaultDataParams> {
    const cachedVaultReport = this.vaultReportByAddress.get(vault);
    if (cachedVaultReport === undefined) {
      return this.getLatestSubmitVaultReportParams(vault);
    } else {
      return cachedVaultReport;
    }
  }
}
