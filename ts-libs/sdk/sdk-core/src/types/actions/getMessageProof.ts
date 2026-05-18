import { MessageProof } from "../message";
import { Hex } from "../misc";

export type GetMessageProofParameters<T> = {
  l2Client: T;
  l2MessageTreeDepth?: number;
  messageHash: Hex;
};

export type GetMessageProofReturnType = MessageProof;
