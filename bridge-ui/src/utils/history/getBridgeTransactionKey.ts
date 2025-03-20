import { BridgeTransaction } from "@/types";

export function getBridgeTransactionKey(transaction: BridgeTransaction): string {
  return `${transaction.bridgingTx}-${transaction.claimingTx}`;
}
