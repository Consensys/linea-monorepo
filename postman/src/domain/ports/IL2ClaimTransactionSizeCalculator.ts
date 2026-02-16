import { MessageProps } from "../message/Message";
import { LineaGasFees } from "../types";

export interface IL2ClaimTransactionSizeCalculator {
  calculateTransactionSize(message: MessageProps & { feeRecipient?: string }, fees: LineaGasFees): Promise<number>;
}
