import { Account, Chain, Client, Transport } from "viem";
import { withdraw, WithdrawParameters, WithdrawReturnType } from "../actions/withdraw";

export type WalletActionsL2 = {
  withdraw: (args: WithdrawParameters) => Promise<WithdrawReturnType>;
};

export function walletActionsL2() {
  return function <
    TChain extends Chain | undefined = Chain | undefined,
    TAccount extends Account | undefined = Account | undefined,
  >(client: Client<Transport, TChain, TAccount>): WalletActionsL2 {
    return {
      withdraw: (args) => withdraw(client, args),
    };
  };
}
