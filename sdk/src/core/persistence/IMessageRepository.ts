import { Message } from "../entities/Message";
import { Direction } from "../enums/MessageEnums";

export interface IMessageRepository<TransactionResponse> {
  insertMessage(message: Message): Promise<void>;
  updateMessage(message: Message): Promise<void>;
  updateMessageByTransactionHash(transactionHash: string, direction: Direction, message: Message): Promise<void>;
  saveMessages(messages: Message[]): Promise<void>;
  deleteMessages(msBeforeNowToDelete: number): Promise<number>;
  getFirstMessageToClaim(
    direction: Direction,
    contractAddress: string,
    currentGasPrice: bigint,
    gasEstimationMargin: number,
    maxRetry: number,
    retryDelay: number,
  ): Promise<Message | null>;
  getLatestMessageSent(direction: Direction, contractAddress: string): Promise<Message | null>;
  getNFirstMessageSent(direction: Direction, limit: number, contractAddress: string): Promise<Message[]>;
  getLastClaimTxNonce(direction: Direction): Promise<number | null>;
  getFirstPendingMessage(direction: Direction): Promise<Message | null>;
  updateMessageWithClaimTxAtomic(
    message: Message,
    nonce: number,
    claimTxResponsePromise: Promise<TransactionResponse>,
  ): Promise<void>;
}
