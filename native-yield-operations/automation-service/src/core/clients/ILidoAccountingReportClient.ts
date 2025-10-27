import { Address } from "viem";
import { UpdateVaultDataParams } from "./contracts/ILazyOracle.js";

export interface ILidoAccountingReportClient {
  getVault(): Address;
  getLatestSubmitVaultReportParams(): Promise<UpdateVaultDataParams>;
  isSimulateSubmitLatestVaultReportSuccessful(): Promise<boolean>;
  submitLatestVaultReport(): Promise<void>;
}
