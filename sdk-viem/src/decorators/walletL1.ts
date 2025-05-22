import { Account, Chain, Client, Transport } from "viem";
import { deposit, DepositParameters, DepositReturnType } from "../actions/deposit";

export type WalletActionsL1 = {
  deposit: (args: DepositParameters) => Promise<DepositReturnType>;
};

export function walletActionsL1() {
  return <
    chain extends Chain | undefined = Chain | undefined,
    account extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, chain, account>,
  ): WalletActionsL1 => ({
    deposit: (args: DepositParameters) => deposit(client, args),
  });
}
