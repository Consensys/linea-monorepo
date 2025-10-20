export interface LidoStakingVaultWithdrawalParams {
  pubkeys: string[];
  amountsGwei: bigint[];
  refundRecipient: string;
}
