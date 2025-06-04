import { Account, Chain, Client, Transport } from "viem";
import { L1WalletClient } from "@consensys/linea-sdk-core";
import { deposit, DepositParameters, DepositReturnType } from "../actions/deposit";
import { StrictFunctionOnly } from "../types/misc";
import { claimOnL1, ClaimOnL1Parameters, ClaimOnL1ReturnType } from "../actions/claimOnL1";

export type WalletActionsL1 = StrictFunctionOnly<
  L1WalletClient,
  {
    deposit: (args: DepositParameters) => Promise<DepositReturnType>;
    claimOnL1: (args: ClaimOnL1Parameters) => Promise<ClaimOnL1ReturnType>;
  }
>;

export function walletActionsL1() {
  return <
    TChain extends Chain | undefined = Chain | undefined,
    TAccount extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, TChain, TAccount>,
  ): WalletActionsL1 => ({
    deposit: (args) => deposit(client, args),
    claimOnL1: (args) => claimOnL1(client, args),
  });
}
