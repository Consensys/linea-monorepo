import { BlockExtraData, BlockTag } from "./block";
import { ExtendedMessage, MessageProof, OnChainMessageStatus } from "./message";
import { Hex } from "./misc";
import { TransactionReceipt } from "./transaction";

export type GetMessageByMessageHashParameters = {
  messageHash: Hex;
};

export type GetMessagesByTransactionHashParameters = {
  transactionHash: Hex;
};

export type GetTransactionReceiptByMessageHashParameters = {
  messageHash: Hex;
};

export interface Provider {
  getMessageByMessageHash(args: GetMessageByMessageHashParameters): Promise<ExtendedMessage>;
  getMessagesByTransactionHash(args: GetMessagesByTransactionHashParameters): Promise<ExtendedMessage[]>;
  getTransactionReceiptByMessageHash(args: GetTransactionReceiptByMessageHashParameters): Promise<TransactionReceipt>;
}

export type GetBlockExtraDataParameters<blockTag extends BlockTag = "latest"> = {
  blockHash?: Hex | undefined;
  blockNumber?: bigint | undefined;
  blockTag?: blockTag | BlockTag | undefined;
};

export type GetL1ToL2MessageStatusParameters = {
  messageHash: Hex;
};

export interface L2Provider extends Provider {
  getBlockExtraData<blockTag extends BlockTag>(args: GetBlockExtraDataParameters<blockTag>): Promise<BlockExtraData>;
  getL1ToL2MessageStatus(args: GetL1ToL2MessageStatusParameters): Promise<OnChainMessageStatus>;
}

export type GetL2ToL1MessageStatusParameters<T> = {
  l2Client: T;
  messageHash: Hex;
};

export type GetMessageProofParameters<T> = {
  l2Client: T;
  messageHash: Hex;
};

export interface L1Provider extends Provider {
  getL2ToL1MessageStatus<T>(args: GetL2ToL1MessageStatusParameters<T>): Promise<OnChainMessageStatus>;
  getMessageProof<T>(parameters: GetMessageProofParameters<T>): Promise<MessageProof>;
}
