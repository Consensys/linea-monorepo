import { Proof } from "./IMerkleTreeService";
import { MessageProps } from "../../../entities/Message";
import { IMessageServiceContract } from "../../../services/contracts/IMessageServiceContract";
import { MessageSent, Overrides } from "../../../types";

export interface ILineaRollupClient extends IMessageServiceContract {
  getMessageProof(messageHash: string): Promise<Proof>;
  estimateClaimGas(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts?: {
      claimViaAddress?: string;
      overrides?: Overrides;
    },
  ): Promise<bigint>;
}
