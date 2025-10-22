import { Address, encodeAbiParameters, Hex } from "viem";

export interface LidoStakingVaultWithdrawalParams {
  pubkeys: Hex[];
  amountsGwei: bigint[];
  refundRecipient: Address;
}

export interface WithdrawalRequests {
  pubkeys: Hex[];
  amountsGwei: bigint[];
}

export function encodeLidoWithdrawalParams(params: LidoStakingVaultWithdrawalParams): Hex {
  return encodeAbiParameters(
    [
      {
        type: "tuple",
        components: [
          { name: "pubkeys", type: "bytes[]" },
          { name: "amountsGwei", type: "uint256[]" },
          { name: "refundRecipient", type: "address" },
        ],
      },
    ],
    [
      {
        pubkeys: params.pubkeys,
        amountsGwei: params.amountsGwei,
        refundRecipient: params.refundRecipient,
      },
    ],
  );
}
