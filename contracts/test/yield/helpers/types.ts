import { SecretKey } from "@chainsafe/blst";

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
}

export interface ValidatorContainer {
  pubkey: string;
  withdrawalCredentials: string;
  effectiveBalance: bigint;
  slashed: boolean;
  activationEligibilityEpoch: bigint;
  activationEpoch: bigint;
  exitEpoch: bigint;
  withdrawableEpoch: bigint;
}

export interface Validator {
  container: ValidatorContainer;
  blsPrivateKey: SecretKey;
}

export interface BeaconBlockHeader {
  slot: number | bigint;
  proposerIndex: number | bigint;
  parentRoot: string;
  stateRoot: string;
  bodyRoot: string;
}

export interface ValidatorWitness {
  proof: string[];
  validatorIndex: bigint;
  effectiveBalance: bigint;
  childBlockTimestamp: bigint;
  slot: bigint;
  proposerIndex: bigint;
  activationEpoch: bigint;
  activationEligibilityEpoch: bigint;
}

export interface EIP4788Witness {
  // Beacon block root
  blockRoot: string;
  // GI First Validator
  gIFirstValidator: string;
  beaconBlockHeader: BeaconBlockHeader;
  witness: ValidatorContainerWitness;
}

export interface ValidatorContainerWitness {
  validatorIndex: bigint;
  validator: ValidatorContainer;
  proof: string[];
}

export interface ClaimMessageWithProofParams {
  proof: string[];
  messageNumber: bigint;
  leafIndex: bigint;
  from: string;
  to: string;
  fee: bigint;
  value: bigint;
  feeRecipient: string;
  merkleRoot: string;
  data: string;
}
