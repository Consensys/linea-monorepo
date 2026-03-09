import { MessageProps } from "../../../entities/Message";
import { IMessageServiceContract } from "../../../services/contracts/IMessageServiceContract";
import { MessageSent, Overrides } from "../../../types";
import { LineaGasFees } from "../IGasProvider";

export interface IL2MessageServiceClient extends IMessageServiceContract {
  encodeClaimMessageTransactionData(message: MessageProps & { feeRecipient?: string }): string;
  estimateClaimGasFees(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    opts?: {
      claimViaAddress?: string;
      overrides?: Overrides;
    },
  ): Promise<LineaGasFees>;
  getContractAddress(): string;
}
