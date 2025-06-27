import { Account, Address, Chain, Client, DeriveChain, Transport } from "viem";
import { withdraw, WithdrawParameters, WithdrawReturnType } from "../actions/withdraw";
import { StrictFunctionOnly } from "../types/misc";
import { L2WalletClient } from "@consensys/linea-sdk-core";
import { claimOnL2, ClaimOnL2Parameters, ClaimOnL2ReturnType } from "../actions/claimOnL2";

export type WalletActionsL2<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = StrictFunctionOnly<
  L2WalletClient,
  {
    withdraw: <
      chainOverride extends Chain | undefined = Chain | undefined,
      derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
    >(
      args: WithdrawParameters<chain, account, chainOverride, derivedChain>,
    ) => Promise<WithdrawReturnType>;
    claimOnL2: <
      chainOverride extends Chain | undefined = Chain | undefined,
      derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
    >(
      args: ClaimOnL2Parameters<chain, account, chainOverride, derivedChain>,
    ) => Promise<ClaimOnL2ReturnType>;
  }
>;

type WalletActionsL2Parameters = {
  l2MessageServiceAddress: Address;
  l2TokenBridgeAddress: Address;
};

export function walletActionsL2(parameters?: WalletActionsL2Parameters) {
  return <
    TChain extends Chain | undefined = Chain | undefined,
    TAccount extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, TChain, TAccount>,
  ): WalletActionsL2 => ({
    withdraw: (args) =>
      withdraw(client, {
        ...args,
        ...(parameters
          ? {
              l2MessageServiceAddress: parameters.l2MessageServiceAddress,
              l2TokenBridgeAddress: parameters.l2TokenBridgeAddress,
            }
          : {}),
      }),
    claimOnL2: (args) =>
      claimOnL2(client, {
        ...args,
        ...(parameters ? { l2MessageServiceAddress: parameters.l2MessageServiceAddress } : {}),
      }),
  });
}
