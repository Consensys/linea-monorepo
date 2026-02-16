import { Message } from "../message/Message";
import { Direction, MessageStatus, TransactionResponse } from "../types";

export interface IMessageDBService {
  insertMessage(message: Message): Promise<void>;

  saveMessages(messages: Message[]): Promise<void>;

  updateMessage(message: Message): Promise<void>;

  deleteMessages(msBeforeNowToDelete: number): Promise<number>;

  getFirstPendingMessage(direction: Direction): Promise<Message | null>;

  getLatestMessageSent(direction: Direction, contractAddress: string): Promise<Message | null>;

  getNFirstMessagesSent(limit: number, contractAddress: string): Promise<Message[]>;

  getNFirstMessagesByStatus(
    status: MessageStatus,
    direction: Direction,
    limit: number,
    contractAddress: string,
  ): Promise<Message[]>;

  getMessageToClaim(
    contractAddress: string,
    gasEstimationMargin: number,
    maxRetry: number,
    retryDelay: number,
  ): Promise<Message | null>;

  getLastClaimTxNonce(direction: Direction): Promise<number | null>;

  updateMessageWithClaimTxAtomic(
    message: Message,
    nonce: number,
    claimTxFn: () => Promise<TransactionResponse>,
  ): Promise<void>;
}
