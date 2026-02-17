import { Message } from "../../../domain/message/Message";
import { Direction } from "../../../domain/types/Direction";
import { MessageStatus } from "../../../domain/types/MessageStatus";

import type { IMessageDBService } from "../../../domain/ports/IMessageDBService";
import type { IMessageRepository } from "../../../domain/ports/IMessageRepository";
import type { TransactionResponse } from "../../../domain/types/BlockchainTypes";

/**
 * DB service for the L1→L2 flow.
 * Claims happen on L2 (Linea), so getMessageToClaim uses L2-specific query logic.
 */
export class LineaMessageDBService implements IMessageDBService {
  constructor(private readonly messageRepository: IMessageRepository) {}

  public async insertMessage(message: Message): Promise<void> {
    return this.messageRepository.insertMessage(message);
  }

  public async saveMessages(messages: Message[]): Promise<void> {
    return this.messageRepository.saveMessages(messages);
  }

  public async updateMessage(message: Message): Promise<void> {
    return this.messageRepository.updateMessage(message);
  }

  public async deleteMessages(msBeforeNowToDelete: number): Promise<number> {
    return this.messageRepository.deleteMessages(msBeforeNowToDelete);
  }

  public async getFirstPendingMessage(direction: Direction): Promise<Message | null> {
    return this.messageRepository.getFirstPendingMessage(direction);
  }

  public async getLatestMessageSent(direction: Direction, contractAddress: string): Promise<Message | null> {
    return this.messageRepository.getLatestMessageSent(direction, contractAddress);
  }

  public async getNFirstMessagesSent(limit: number, contractAddress: string): Promise<Message[]> {
    return this.messageRepository.getNFirstMessagesByStatus(
      MessageStatus.SENT,
      Direction.L1_TO_L2,
      limit,
      contractAddress,
    );
  }

  public async getNFirstMessagesByStatus(
    status: MessageStatus,
    direction: Direction,
    limit: number,
    contractAddress: string,
  ): Promise<Message[]> {
    return this.messageRepository.getNFirstMessagesByStatus(status, direction, limit, contractAddress);
  }

  public async getMessageToClaim(
    contractAddress: string,
    _gasEstimationMargin: number,
    maxRetry: number,
    retryDelay: number,
  ): Promise<Message | null> {
    return this.messageRepository.getFirstMessageToClaimOnL2(
      Direction.L1_TO_L2,
      contractAddress,
      [MessageStatus.TRANSACTION_SIZE_COMPUTED, MessageStatus.FEE_UNDERPRICED],
      maxRetry,
      retryDelay,
    );
  }

  public async getLastClaimTxNonce(direction: Direction): Promise<number | null> {
    return this.messageRepository.getLastClaimTxNonce(direction);
  }

  public async updateMessageWithClaimTxAtomic(
    message: Message,
    nonce: number,
    claimTxFn: () => Promise<TransactionResponse>,
  ): Promise<void> {
    await this.messageRepository.updateMessageWithClaimTxAtomic(message, nonce, claimTxFn);
  }
}
