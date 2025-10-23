import { Address, TransactionReceipt, BaseError, ContractFunctionRevertedError } from "viem";
import { ILidoAccountingReportClient } from "../core/clients/ILidoAccountingReportClient.js";
import { ILazyOracle, UpdateVaultDataParams } from "../core/services/contracts/ILazyOracle.js";
import { getReportProofByVault } from "@lidofinance/lsv-cli";
import { ILogger } from "@consensys/linea-shared-utils";

export class LidoAccountingReportClient implements ILidoAccountingReportClient {
  private latestSubmitVaultReportParams?: UpdateVaultDataParams;

  constructor(
    private readonly lazyOracleContractClient: ILazyOracle<TransactionReceipt>,
    private readonly ipfsGatewayUrl: string,
    private readonly logger: ILogger,
    private readonly vault: Address,
  ) {}

  async getLatestSubmitVaultReportParams(): Promise<UpdateVaultDataParams> {
    const latestReportData = await this.lazyOracleContractClient.latestReportData();

    const reportProof = await getReportProofByVault({
      vault: this.vault,
      cid: latestReportData.reportCid,
      gateway: this.ipfsGatewayUrl,
    });

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
        console.error("❌ Reverted:", err.shortMessage);
        console.error("Reason:", err.data?.errorName || err.message);
      } else if (err instanceof BaseError) {
        console.error("⚠️ Other error:", err.shortMessage);
      } else {
        console.error("Unexpected error:", err);
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
