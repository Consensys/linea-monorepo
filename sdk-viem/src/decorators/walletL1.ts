import { Account, Chain, Client, Transport } from "viem";
import { deposit, DepositParameters, DepositReturnType } from "../actions/deposit";

export type WalletActionsL1 = {
  deposit: (args: DepositParameters) => Promise<DepositReturnType>;
};

export function walletActionsL1() {
  return function <
    TChain extends Chain | undefined = Chain | undefined,
    TAccount extends Account | undefined = Account | undefined,
  >(client: Client<Transport, TChain, TAccount>): WalletActionsL1 {
    return {
      deposit: (args: DepositParameters) => deposit(client, args),
    };
  };
}
