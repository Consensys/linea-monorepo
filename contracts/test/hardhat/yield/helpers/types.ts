export interface YieldManagerInitializationData {
  pauseTypeRoles: { pauseType: number; role: string }[];
  unpauseTypeRoles: { pauseType: number; role: string }[];
  roleAddresses: { addressWithRole: string; role: string }[];
  initialL2YieldRecipients: string[];
  defaultAdmin: string;
  initialMinimumWithdrawalReservePercentageBps: number; // uint16 on-chain
  initialTargetWithdrawalReservePercentageBps: number; // uint16 on-chain
  initialMinimumWithdrawalReserveAmount: bigint; // uint256 on-chain
  initialTargetWithdrawalReserveAmount: bigint; // uint256 on-chain
}

export interface YieldProviderRegistration {
  yieldProviderVendor: number;
  primaryEntrypoint: string;
  ossifiedEntrypoint: string;
  receiveCaller: string;
}
