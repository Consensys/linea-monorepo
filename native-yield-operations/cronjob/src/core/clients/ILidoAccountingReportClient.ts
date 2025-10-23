import { UpdateVaultDataParams } from "../services/contracts/ILazyOracle.js";

export interface ILidoAccountingReportClient {
  // i.) Get latest vault report CID
  // ii.) Return early if latest vault report already submitted
  // iii.) Get accounting report from IPFS
  // iv.) Compute params for updateVaultData
  // v.) Submit tx
  // vi.) Return success
  getLatestSubmitVaultReportParams(): Promise<UpdateVaultDataParams>;
  isSimulateSubmitLatestVaultReportSuccessful(): Promise<boolean>;
  submitLatestVaultReport(): Promise<void>;
}
