import { LineaGasFees } from "../../clients/blockchain/IGasProvider";
import { MessageProps } from "../../entities/Message";

import type { Address } from "../../types/hex";

export interface IL2ClaimTransactionSizeCalculator {
  calculateTransactionSize(message: MessageProps & { feeRecipient?: Address }, fees: LineaGasFees): Promise<number>;
}
