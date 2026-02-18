import { MessageProps } from "../message/Message";

import type { LineaGasFees } from "../types/blockchain";

export interface IL2ClaimTransactionSizeCalculator {
  calculateTransactionSize(message: MessageProps & { feeRecipient?: string }, fees: LineaGasFees): Promise<number>;
}
