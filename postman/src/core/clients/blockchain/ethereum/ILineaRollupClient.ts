import { Proof } from "./IMerkleTreeService";
import { MessageProps } from "../../../entities/Message";
import { IMessageServiceContract } from "../../../services/contracts/IMessageServiceContract";
import { Address, Hash, MessageSent, Overrides } from "../../../types";

export interface ILineaRollupClient extends IMessageServiceContract {
  getMessageProof(messageHash: Hash): Promise<Proof>;
  estimateClaimGas(
    message: (MessageSent | MessageProps) & { feeRecipient?: Address; messageBlockNumber?: number },
    opts?: {
      claimViaAddress?: Address;
      overrides?: Overrides;
    },
  ): Promise<bigint>;
}
