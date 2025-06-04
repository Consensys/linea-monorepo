import { Account, Chain, Client, Transport } from "viem";
import { withdraw, WithdrawParameters, WithdrawReturnType } from "../actions/withdraw";
import { StrictFunctionOnly } from "../types/misc";
import { L2WalletClient } from "@consensys/linea-sdk-core";
import { claimOnL2, ClaimOnL2Parameters, ClaimOnL2ReturnType } from "../actions/claimOnL2";

export type WalletActionsL2 = StrictFunctionOnly<
  L2WalletClient,
  {
    withdraw: (args: WithdrawParameters) => Promise<WithdrawReturnType>;
    claimOnL2: (args: ClaimOnL2Parameters) => Promise<ClaimOnL2ReturnType>;
  }
>;

export function walletActionsL2() {
  return <
    TChain extends Chain | undefined = Chain | undefined,
    TAccount extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, TChain, TAccount>,
  ): WalletActionsL2 => ({
    withdraw: (args) => withdraw(client, args),
    claimOnL2: (args) => claimOnL2(client, args),
  });
}
