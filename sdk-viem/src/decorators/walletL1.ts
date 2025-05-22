import { Account, Chain, Client, Transport } from "viem";
import { DepositParameters } from "../actions/deposit";

export type WalletActionsL1 = {
  deposit: (args: DepositParameters) => void;
};

export function walletActionsL1() {
  return <
    chain extends Chain | undefined = Chain | undefined,
    account extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, chain, account>,
  ): WalletActionsL1 => ({
    deposit: (args: DepositParameters) => {
      console.log("Deposit called with args:", client, args);
    },
  });
}
