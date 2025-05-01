import {
  EventSubscriber,
  EntitySubscriberInterface,
  InsertEvent,
  UpdateEvent,
  RemoveEvent,
  TransactionCommitEvent,
} from "typeorm";
import { Direction } from "@consensys/linea-sdk";
import { MessageEntity } from "../entities/Message.entity";
import { IMessageMetricsUpdater, MessagesMetricsAttributes } from "../../../../core/metrics";
import { ILogger } from "../../../../core/utils/logging/ILogger";
import { MessageStatus } from "../../../../core/enums";

@EventSubscriber()
export class MessageStatusSubscriber implements EntitySubscriberInterface<MessageEntity> {
  constructor(
    private readonly messageMetricsUpdater: IMessageMetricsUpdater,
    private readonly logger: ILogger,
  ) {}

  listenTo() {
    return MessageEntity;
  }

  async afterInsert(event: InsertEvent<MessageEntity>): Promise<void> {
    const { status, direction, isForSponsorship } = event.entity;
    this.updateMessageMetricsOnInsert(status, direction, isForSponsorship);
  }

  async afterUpdate(event: UpdateEvent<MessageEntity>): Promise<void> {
    if (!event.entity || !event.databaseEntity) return;

    const prevStatus = event.databaseEntity.status;
    const newStatus = event.entity.status;
    const prevIsForSponsorship = event.databaseEntity.isForSponsorship;
    const newIsForSponsorship = event.entity.isForSponsorship;
    const direction = event.databaseEntity.direction;

    if (prevStatus !== newStatus || prevIsForSponsorship !== newIsForSponsorship) {
      await this.swapMessageAttributes(
        { status: prevStatus, direction, isForSponsorship: prevIsForSponsorship },
        { status: newStatus, direction, isForSponsorship: newIsForSponsorship },
      );
    }
  }

  async afterRemove(event: RemoveEvent<MessageEntity>): Promise<void> {
    if (event.entity) {
      await this.updateMessageMetricsOnRemove(
        event.databaseEntity.status,
        event.databaseEntity.direction,
        event.databaseEntity.isForSponsorship,
      );
    }
  }

  async afterTransactionCommit(event: TransactionCommitEvent): Promise<void> {
    const updatedEntity = event.queryRunner?.data?.updatedEntity;
    if (updatedEntity) {
      await this.swapMessageAttributes(
        {
          status: updatedEntity.previousStatus,
          direction: updatedEntity.direction,
          isForSponsorship: updatedEntity.isForSponsorship,
        },
        {
          status: updatedEntity.newStatus,
          direction: updatedEntity.direction,
          isForSponsorship: updatedEntity.isForSponsorship,
        },
      );
    }
  }

  private async updateMessageMetricsOnInsert(
    messageStatus: MessageStatus,
    messageDirection: Direction,
    isForSponsorship: boolean,
  ): Promise<void> {
    try {
      const prevGaugeValue = await this.messageMetricsUpdater.getMessageCount({
        status: messageStatus,
        direction: messageDirection,
        isForSponsorship: isForSponsorship,
      });

      if (prevGaugeValue === undefined) {
        return;
      }

      this.messageMetricsUpdater.incrementMessageCount({
        status: messageStatus,
        direction: messageDirection,
        isForSponsorship: isForSponsorship,
      });
    } catch (error) {
      this.logger.error("Failed to update metrics:", error);
    }
  }

  private async updateMessageMetricsOnRemove(
    messageStatus: MessageStatus,
    messageDirection: Direction,
    isForSponsorship: boolean,
  ): Promise<void> {
    try {
      const prevGaugeValue = await this.messageMetricsUpdater.getMessageCount({
        status: messageStatus,
        direction: messageDirection,
        isForSponsorship: isForSponsorship,
      });

      if (prevGaugeValue && prevGaugeValue > 0) {
        this.messageMetricsUpdater.decrementMessageCount({
          status: messageStatus,
          direction: messageDirection,
          isForSponsorship: isForSponsorship,
        });
      }
    } catch (error) {
      this.logger.error("Failed to update metrics:", error);
    }
  }

  private async swapMessageAttributes(
    previousMessageAttributes: MessagesMetricsAttributes,
    nextMessageAttributes: MessagesMetricsAttributes,
  ): Promise<void> {
    try {
      const [prevVal, newVal] = await Promise.all([
        this.messageMetricsUpdater.getMessageCount(previousMessageAttributes),
        this.messageMetricsUpdater.getMessageCount(nextMessageAttributes),
      ]);

      if (prevVal && prevVal > 0) {
        this.messageMetricsUpdater.decrementMessageCount(previousMessageAttributes);
      }

      if (newVal !== undefined) {
        this.messageMetricsUpdater.incrementMessageCount(nextMessageAttributes);
      }
    } catch (error) {
      this.logger.error("Metrics swap failed:", error);
    }
  }
}
