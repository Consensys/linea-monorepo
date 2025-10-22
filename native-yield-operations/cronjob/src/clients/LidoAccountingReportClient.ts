import { Address, TransactionReceipt, BaseError, ContractFunctionRevertedError } from "viem";
import { ILidoAccountingReportClient } from "../core/clients/ILidoAccountingReportClient";
import { ILazyOracle, UpdateVaultDataParams } from "../core/services/contracts/ILazyOracle";
import { getReportProofByVault } from "@lidofinance/lsv-cli";
import { ILogger } from "ts-libs/linea-shared-utils";

export class LidoAccountingReportClient implements ILidoAccountingReportClient {
  constructor(
    private readonly lazyOracleContractClient: ILazyOracle<TransactionReceipt>,
    private readonly ipfsGatewayUrl: string,
    private readonly logger: ILogger,
  ) {}

  async getSubmitVaultReportParams(vaultAddress: Address): Promise<UpdateVaultDataParams> {
    const latestReportData = await this.lazyOracleContractClient.latestReportData();

    const reportProof = await getReportProofByVault({
      vault: vaultAddress,
      cid: latestReportData.reportCid,
      gateway: this.ipfsGatewayUrl,
    });

    return {
      vault: vaultAddress,
      totalValue: BigInt(reportProof.data.totalValueWei),
      cumulativeLidoFees: BigInt(reportProof.data.fee),
      liabilityShares: BigInt(reportProof.data.liabilityShares),
      maxLiabilityShares: BigInt(reportProof.data.maxLiabilityShares),
      slashingReserve: BigInt(reportProof.data.slashingReserve),
      proof: reportProof.proof,
    };
  }

  // i.) Get latest vault report CID
  // ii.) Return early if latest vault report already submitted
  // iii.) Get accounting report from IPFS
  // iv.) Compute params for updateVaultData
  // v.) Submit tx
  // vi.) Return success
  // Return true is simulated succeeded, false otherwise
  async isSimulateSubmitLatestVaultReportSuccessful(params: UpdateVaultDataParams): Promise<boolean> {
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

  async submitLatestVaultReport(params: UpdateVaultDataParams): Promise<void> {
    await this.lazyOracleContractClient.updateVaultData(params);
  }
}
