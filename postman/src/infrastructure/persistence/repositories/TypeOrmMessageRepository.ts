/* eslint-disable @typescript-eslint/no-explicit-any */
import { Brackets, DataSource, Repository } from "typeorm";

import { DatabaseAccessError } from "../../../domain/errors";
import { Message } from "../../../domain/message/Message";
import { IMessageRepository } from "../../../domain/ports/IMessageRepository";
import {
  DatabaseErrorType,
  DatabaseRepoName,
  Direction,
  MessageStatus,
  TransactionResponse,
} from "../../../domain/types";
import { subtractSeconds } from "../../../domain/utils";
import { MessageEntity } from "../entities/MessageEntity";
import { mapMessageEntityToMessage, mapMessageToMessageEntity } from "../mappers/MessageMapper";

export class TypeOrmMessageRepository extends Repository<MessageEntity> implements IMessageRepository {
  constructor(readonly dataSource: DataSource) {
    super(MessageEntity, dataSource.createEntityManager());
  }

  async insertMessage(message: Message): Promise<void> {
    try {
      const messageInDb = await this.findOneBy({
        messageHash: message.messageHash,
        direction: message.direction,
      });
      if (!messageInDb) {
        await this.manager.save(MessageEntity, mapMessageToMessageEntity(message));
      }
    } catch (err: any) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Insert, err, message);
    }
  }

  async updateMessage(message: Message): Promise<void> {
    try {
      const messageInDb = await this.findOneBy({
        messageHash: message.messageHash,
        direction: message.direction,
      });
      if (messageInDb) {
        await this.manager.save(MessageEntity, mapMessageToMessageEntity(message));
      }
    } catch (err: any) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Update, err, message);
    }
  }

  async saveMessages(messages: Message[]): Promise<void> {
    try {
      await this.manager.save(MessageEntity, messages.map(mapMessageToMessageEntity));
    } catch (err: any) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Insert, err);
    }
  }

  async deleteMessages(msBeforeNowToDelete: number): Promise<number> {
    try {
      const d = new Date();
      d.setTime(d.getTime() - msBeforeNowToDelete);
      const formattedDateStr = d.toISOString().replace("T", " ").replace("Z", "");
      const deleteResult = await this.createQueryBuilder("message")
        .delete()
        .where("message.status IN(:...statuses)", {
          statuses: [
            MessageStatus.CLAIMED_SUCCESS,
            MessageStatus.CLAIMED_REVERTED,
            MessageStatus.EXCLUDED,
            MessageStatus.ZERO_FEE,
          ],
        })
        .andWhere("message.updated_at < :updated_before", { updated_before: formattedDateStr })
        .execute();
      return deleteResult.affected ?? 0;
    } catch (err: any) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Delete, err);
    }
  }

  async getFirstMessageToClaimOnL1(
    direction: Direction,
    contractAddress: string,
    currentGasPrice: bigint,
    gasEstimationMargin: number,
    maxRetry: number,
    retryDelay: number,
  ): Promise<Message | null> {
    try {
      const message = await this.createQueryBuilder("message")
        .where("message.direction = :direction", { direction })
        .andWhere("message.messageContractAddress = :contractAddress", { contractAddress })
        .andWhere("message.status IN(:...statuses)", {
          statuses: [MessageStatus.ANCHORED, MessageStatus.FEE_UNDERPRICED],
        })
        .andWhere("message.claimNumberOfRetry < :maxRetry", { maxRetry })
        .andWhere(
          new Brackets((qb) => {
            qb.where("message.claimLastRetriedAt IS NULL").orWhere("message.claimLastRetriedAt < :lastRetriedDate", {
              lastRetriedDate: subtractSeconds(new Date(), retryDelay).toISOString(),
            });
          }),
        )
        .andWhere(
          new Brackets((qb) => {
            qb.where("message.claimGasEstimationThreshold > :threshold", {
              threshold: parseFloat(currentGasPrice.toString()) * gasEstimationMargin,
            }).orWhere("message.claimGasEstimationThreshold IS NULL");
          }),
        )
        .orderBy("CAST(message.status as CHAR)", "ASC")
        .addOrderBy("message.claimGasEstimationThreshold", "DESC")
        .addOrderBy("message.sentBlockNumber", "ASC")
        .getOne();

      return message ? mapMessageEntityToMessage(message) : null;
    } catch (err: any) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err);
    }
  }

  async getFirstMessageToClaimOnL2(
    direction: Direction,
    contractAddress: string,
    messageStatuses: MessageStatus[],
    maxRetry: number,
    retryDelay: number,
  ): Promise<Message | null> {
    try {
      const message = await this.createQueryBuilder("message")
        .where("message.direction = :direction", { direction })
        .andWhere("message.messageContractAddress = :contractAddress", { contractAddress })
        .andWhere("message.status IN(:...statuses)", {
          statuses: messageStatuses,
        })
        .andWhere("message.claimNumberOfRetry < :maxRetry", { maxRetry })
        .andWhere(
          new Brackets((qb) => {
            qb.where("message.claimLastRetriedAt IS NULL").orWhere("message.claimLastRetriedAt < :lastRetriedDate", {
              lastRetriedDate: subtractSeconds(new Date(), retryDelay).toISOString(),
            });
          }),
        )
        .orderBy("CAST(message.status as CHAR)", "DESC")
        .addOrderBy("CAST(message.fee AS numeric)", "DESC")
        .addOrderBy("message.sentBlockNumber", "ASC")
        .getOne();

      return message ? mapMessageEntityToMessage(message) : null;
    } catch (err: any) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err);
    }
  }

  async getLatestMessageSent(direction: Direction, contractAddress: string): Promise<Message | null> {
    try {
      const pendingMessages = await this.find({
        where: {
          direction,
          messageContractAddress: contractAddress,
        },
        take: 1,
        order: {
          createdAt: "DESC",
        },
      });

      if (pendingMessages.length === 0) return null;

      return mapMessageEntityToMessage(pendingMessages[0]);
    } catch (err: any) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err);
    }
  }

  async getNFirstMessagesByStatus(
    status: MessageStatus,
    direction: Direction,
    limit: number,
    contractAddress: string,
  ): Promise<Message[]> {
    try {
      const messages = await this.find({
        where: {
          direction,
          status,
          messageContractAddress: contractAddress,
        },
        take: limit,
        order: {
          sentBlockNumber: "ASC",
        },
      });
      return messages.map(mapMessageEntityToMessage);
    } catch (err: any) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err);
    }
  }

  async getLastClaimTxNonce(direction: Direction): Promise<number | null> {
    try {
      const message = await this.createQueryBuilder("message")
        .select("MAX(message.claimTxNonce)", "lastTxNonce")
        .where("message.direction = :direction", { direction })
        .getRawOne();

      if (message.lastTxNonce === null || message.lastTxNonce === undefined) {
        return null;
      }

      return message.lastTxNonce;
    } catch (err: any) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err);
    }
  }

  async getFirstPendingMessage(direction: Direction): Promise<Message | null> {
    try {
      const message = await this.createQueryBuilder("message")
        .where("message.direction = :direction", { direction })
        .andWhere("message.status = :status", { status: MessageStatus.PENDING })
        .orderBy("message.claimTxNonce", "ASC")
        .getOne();

      return message ? mapMessageEntityToMessage(message) : null;
    } catch (err: any) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err);
    }
  }

  async updateMessageWithClaimTxAtomic(
    message: Message,
    nonce: number,
    claimTxFn: () => Promise<TransactionResponse>,
  ): Promise<void> {
    await this.manager.transaction(async (entityManager) => {
      await entityManager.update(
        MessageEntity,
        { messageHash: message.messageHash, direction: message.direction },
        {
          claimTxNonce: nonce,
          status: MessageStatus.PENDING,
          ...(message.status === MessageStatus.FEE_UNDERPRICED
            ? { claimNumberOfRetry: message.claimNumberOfRetry + 1, claimLastRetriedAt: new Date() }
            : {}),
        },
      );

      const claimTxCreationDate = new Date();
      const tx = await claimTxFn();

      await entityManager.update(
        MessageEntity,
        { messageHash: message.messageHash, direction: message.direction },
        {
          claimTxCreationDate,
          claimTxGasLimit: Number(tx.gasLimit),
          claimTxMaxFeePerGas: tx.maxFeePerGas ?? undefined,
          claimTxMaxPriorityFeePerGas: tx.maxPriorityFeePerGas ?? undefined,
          claimTxHash: tx.hash,
        },
      );

      entityManager.queryRunner!.data.updatedEntity = {
        previousStatus: message.status,
        newStatus: MessageStatus.PENDING,
        direction: message.direction,
      };
    });
  }
}
