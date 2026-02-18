import { MessageProps } from "../message/Message";

import type { GasFees, LineaGasFees, MessageSent } from "../types/blockchain";

export interface IClaimGasEstimator {
  estimateClaimGasFees(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts?: { claimViaAddress?: string },
  ): Promise<GasFees | LineaGasFees>;
}
