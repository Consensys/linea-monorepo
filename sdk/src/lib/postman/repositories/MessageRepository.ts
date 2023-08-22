import { Brackets, DataSource, Repository } from "typeorm";
import { BigNumber } from "ethers";
import { MessageEntity } from "../entity/Message.entity";
import { DatabaseAccessError } from "../utils/errors";
import { DatabaseErrorType, DatabaseRepoName, Direction, MessageStatus } from "../utils/enums";
import { MessageInDb } from "../utils/types";
import { mapMessageEntityToMessage, mapMessageToMessageEntity } from "../utils/mappers";
import { subtractSeconds } from "../utils/helpers";

export class MessageRepository extends Repository<MessageEntity> {
  constructor(readonly dataSource: DataSource) {
    super(MessageEntity, dataSource.createEntityManager());
  }

  async findByMessageHash(message: MessageInDb, direction: Direction): Promise<MessageInDb | null> {
    try {
      const messageInDb = await this.findOneBy({
        messageHash: message.messageHash,
        direction,
      });
      if (!messageInDb) {
        return null;
      }
      return mapMessageEntityToMessage(messageInDb);
    } catch (err) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err, message);
    }
  }

  async insertMessage(message: MessageInDb): Promise<void> {
    try {
      const messageInDb = await this.findOneBy({
        messageHash: message.messageHash,
        direction: message.direction,
      });
      if (!messageInDb) {
        await this.manager.save(MessageEntity, mapMessageToMessageEntity(message));
      }
    } catch (err) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Insert, err, message);
    }
  }

  async updateMessage(
    messageHash: string,
    direction: Direction,
    messagePropertiesToUpdate: Partial<MessageInDb>,
  ): Promise<void> {
    let messageInDb: MessageEntity | null;
    try {
      const messageInDb = await this.findOneBy({
        messageHash,
        direction,
      });
      if (messageInDb) {
        await this.save(mapMessageToMessageEntity({ ...messageInDb, ...messagePropertiesToUpdate }));
      }
    } catch (err) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Update, err, {
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        ...messageInDb!,
        ...messagePropertiesToUpdate,
      });
    }
  }

  async updateMessageByTransactionHash(
    transactionHash: string,
    direction: Direction,
    messagePropertiesToUpdate: Partial<MessageInDb>,
  ): Promise<void> {
    try {
      const messageInDb = await this.findOneBy({
        claimTxHash: transactionHash,
        direction,
      });
      if (messageInDb) {
        await this.manager.save(
          MessageEntity,
          mapMessageToMessageEntity({ ...messageInDb, ...messagePropertiesToUpdate }),
        );
      }
    } catch (err) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Update, err);
    }
  }

  async saveMessages(messages: MessageInDb[]): Promise<void> {
    try {
      await this.manager.save(MessageEntity, messages.map(mapMessageToMessageEntity));
    } catch (err) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Insert, err);
    }
  }

  async getFirstMessageToClaim(
    direction: Direction,
    contractAddress: string,
    currentGasPrice: BigNumber,
    gasEstimationMargin: number,
    maxRetry: number,
    retryDelay: number,
  ): Promise<MessageInDb | null> {
    try {
      return await this.createQueryBuilder("message")
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
    } catch (err) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err);
    }
  }

  async getLatestMessageSent(direction: Direction, contractAddress: string): Promise<MessageInDb | null> {
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

      return pendingMessages[0];
    } catch (err) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err);
    }
  }

  async getNFirstMessageSent(direction: Direction, limit: number, contractAddress: string): Promise<MessageInDb[]> {
    try {
      return await this.find({
        where: {
          direction,
          status: MessageStatus.SENT,
          messageContractAddress: contractAddress,
        },
        take: limit,
        order: {
          sentBlockNumber: "ASC",
        },
      });
    } catch (err) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err);
    }
  }

  async getLastTxNonce(direction: Direction): Promise<number | null> {
    try {
      const message = await this.createQueryBuilder("message")
        .select("MAX(message.claimTxNonce)", "lastTxNonce")
        .where("message.direction = :direction", { direction })
        .getRawOne();

      if (!message.lastTxNonce) {
        return null;
      }

      return message.lastTxNonce;
    } catch (err) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err);
    }
  }

  async getFirstPendingMessage(direction: Direction): Promise<MessageInDb | null> {
    try {
      const message = await this.createQueryBuilder("message")
        .where("message.direction = :direction", { direction })
        .andWhere("message.status = :status", { status: MessageStatus.PENDING })
        .orderBy("message.claimTxNonce", "ASC")
        .getOne();

      return message;
    } catch (err) {
      throw new DatabaseAccessError(DatabaseRepoName.MessageRepository, DatabaseErrorType.Read, err);
    }
  }
}
