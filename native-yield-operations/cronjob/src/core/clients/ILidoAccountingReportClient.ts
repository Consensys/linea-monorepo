export interface ILidoAccountingReportClient {
  // i.) Get latest vault report CID
  // ii.) Return early if latest vault report already submitted
  // iii.) Get accounting report from IPFS
  // iv.) Compute params for updateVaultData
  // v.) Submit tx
  // vi.) Return success
  submitLatestVaultReport(): Promise<void>;
}
