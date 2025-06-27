import { Account, Address, Chain, Client, DeriveChain, Transport } from "viem";
import { L1WalletClient } from "@consensys/linea-sdk-core";
import { deposit, DepositParameters, DepositReturnType } from "../actions/deposit";
import { StrictFunctionOnly } from "../types/misc";
import { claimOnL1, ClaimOnL1Parameters, ClaimOnL1ReturnType } from "../actions/claimOnL1";

export type WalletActionsL1<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = StrictFunctionOnly<
  L1WalletClient,
  {
    deposit: <
      chainOverride extends Chain | undefined = Chain | undefined,
      chainL2 extends Chain | undefined = Chain | undefined,
      accountL2 extends Account | undefined = Account | undefined,
      derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
    >(
      args: DepositParameters<chain, account, chainOverride, chainL2, accountL2, derivedChain>,
    ) => Promise<DepositReturnType>;
    claimOnL1: <
      chainOverride extends Chain | undefined = Chain | undefined,
      derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
    >(
      args: ClaimOnL1Parameters<chain, account, chainOverride, derivedChain>,
    ) => Promise<ClaimOnL1ReturnType>;
  }
>;

export type WalletActionsL1Parameters = {
  lineaRollupAddress: Address;
  l2MessageServiceAddress: Address;
  l1TokenBridgeAddress: Address;
  l2TokenBridgeAddress: Address;
};

export function walletActionsL1(parameters?: WalletActionsL1Parameters) {
  return <
    TChain extends Chain | undefined = Chain | undefined,
    TAccount extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, TChain, TAccount>,
  ): WalletActionsL1 => ({
    deposit: (args) =>
      deposit(client, {
        ...args,
        ...(parameters
          ? {
              lineaRollupAddress: parameters.lineaRollupAddress,
              l2MessageServiceAddress: parameters.l2MessageServiceAddress,
              l1TokenBridgeAddress: parameters.l1TokenBridgeAddress,
              l2TokenBridgeAddress: parameters.l2TokenBridgeAddress,
            }
          : {}),
      }),
    claimOnL1: (args) =>
      claimOnL1(client, { ...args, ...(parameters ? { lineaRollupAddress: parameters.lineaRollupAddress } : {}) }),
  });
}
