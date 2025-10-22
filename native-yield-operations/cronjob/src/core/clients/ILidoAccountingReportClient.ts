import { Address } from "viem";
import { UpdateVaultDataParams } from "../services/contracts/ILazyOracle";

export interface ILidoAccountingReportClient {
  // i.) Get latest vault report CID
  // ii.) Return early if latest vault report already submitted
  // iii.) Get accounting report from IPFS
  // iv.) Compute params for updateVaultData
  // v.) Submit tx
  // vi.) Return success
  getSubmitVaultReportParams(vaultAddress: Address): Promise<UpdateVaultDataParams>;
  isSimulateSubmitLatestVaultReportSuccessful(params: UpdateVaultDataParams): Promise<boolean>;
  submitLatestVaultReport(params: UpdateVaultDataParams): Promise<void>;
}
