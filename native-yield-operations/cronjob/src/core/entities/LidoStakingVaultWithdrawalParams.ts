import { Address, encodeAbiParameters, Hex } from "viem";

export interface WithdrawalRequests {
  pubkeys: Hex[];
  amountsGwei: bigint[];
}

export interface LidoStakingVaultWithdrawalParams extends WithdrawalRequests {
  refundRecipient: Address;
}

export function encodeLidoWithdrawalParams(params: LidoStakingVaultWithdrawalParams): Hex {
  return encodeAbiParameters(
    [
      {
        type: "tuple",
        components: [
          { name: "pubkeys", type: "bytes[]" },
          { name: "amounts", type: "uint64[]" },
          { name: "refundRecipient", type: "address" },
        ],
      },
    ],
    [
      {
        pubkeys: params.pubkeys,
        amounts: params.amountsGwei,
        refundRecipient: params.refundRecipient,
      },
    ],
  );
}
