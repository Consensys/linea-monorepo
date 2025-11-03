import { Address } from "viem";
import { UpdateVaultDataParams } from "./contracts/ILazyOracle.js";

export interface ILidoAccountingReportClient {
  getLatestSubmitVaultReportParams(vault: Address): Promise<UpdateVaultDataParams>;
  isSimulateSubmitLatestVaultReportSuccessful(vault: Address): Promise<boolean>;
  submitLatestVaultReport(vault: Address): Promise<void>;
}
