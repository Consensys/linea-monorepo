import { Message } from "../../../domain/message/Message";
import { Direction } from "../../../domain/types/Direction";
import { MessageStatus } from "../../../domain/types/MessageStatus";

import type { IGasProvider } from "../../../domain/ports/IGasProvider";
import type { IMessageDBService } from "../../../domain/ports/IMessageDBService";
import type { IMessageRepository } from "../../../domain/ports/IMessageRepository";
import type { TransactionResponse } from "../../../domain/types/BlockchainTypes";

/**
 * DB service for the L2→L1 flow.
 * Claims happen on L1 (Ethereum), so getMessageToClaim uses L1-specific query logic
 * (requires current gas price from the gas provider).
 */
export class EthereumMessageDBService implements IMessageDBService {
  constructor(
    private readonly gasProvider: IGasProvider,
    private readonly messageRepository: IMessageRepository,
  ) {}

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
      Direction.L2_TO_L1,
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
    gasEstimationMargin: number,
    maxRetry: number,
    retryDelay: number,
  ): Promise<Message | null> {
    const { maxFeePerGas } = await this.gasProvider.getGasFees();
    return this.messageRepository.getFirstMessageToClaimOnL1(
      Direction.L2_TO_L1,
      contractAddress,
      maxFeePerGas,
      gasEstimationMargin,
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
