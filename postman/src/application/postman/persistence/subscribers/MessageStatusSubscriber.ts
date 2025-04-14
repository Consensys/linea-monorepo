import {
  EventSubscriber,
  EntitySubscriberInterface,
  InsertEvent,
  UpdateEvent,
  RemoveEvent,
  TransactionCommitEvent,
} from "typeorm";
import { MessageEntity } from "../entities/Message.entity";
import { IMetricService } from "../../../../core/metrics/IMetricService";
import { MessageStatus } from "../../../../core/enums";
import { ILogger } from "../../../../core/utils/logging/ILogger";

@EventSubscriber()
export class MessageStatusSubscriber implements EntitySubscriberInterface<MessageEntity> {
  constructor(
    private readonly metricsService: IMetricService,
    private readonly logger: ILogger,
  ) {}

  listenTo() {
    return MessageEntity;
  }

  async afterInsert(event: InsertEvent<MessageEntity>): Promise<void> {
    await this.updateMessageMetricsOnInsert(event);
  }

  async afterUpdate(event: UpdateEvent<MessageEntity>): Promise<void> {
    if (event.entity && event.databaseEntity) {
      const previousStatus = event.databaseEntity.status;
      const newStatus = event.entity.status;
      if (previousStatus !== newStatus) {
        await this.updateMessageMetricsOnUpdate(event);
      }
    }
  }

  async afterRemove(event: RemoveEvent<MessageEntity>): Promise<void> {
    if (event.entity) {
      await this.updateMessageMetricsOnRemove(event);
    }
  }

  async afterTransactionCommit(event: TransactionCommitEvent): Promise<void> {
    const updatedEntity = event.queryRunner?.data?.updatedEntity;
    if (updatedEntity) {
      const [previousGauge, gauge] = await Promise.all([
        this.metricsService.getGaugeValue("postman_messages_current", {
          status: updatedEntity.previousStatus,
          direction: updatedEntity.direction,
        }),
        this.metricsService.getGaugeValue("postman_messages_current", {
          status: updatedEntity.newStatus,
          direction: updatedEntity.direction,
        }),
      ]);

      if (previousGauge && previousGauge > 0) {
        this.metricsService.decrementGauge("postman_messages_current", 1, {
          status: updatedEntity.previousStatus,
          direction: updatedEntity.direction,
        });
      }

      if (gauge || gauge === 0) {
        this.metricsService.incrementGauge("postman_messages_current", 1, {
          status: updatedEntity.newStatus,
          direction: updatedEntity.direction,
        });
      }
    }
  }

  private async updateMessageMetricsOnInsert(event: InsertEvent<MessageEntity>): Promise<void> {
    try {
      const messageStatus = event.entity.status;
      const messageDirection = event.entity.direction;

      const gauge = await this.metricsService.getGaugeValue("postman_messages_current", {
        status: messageStatus,
        direction: messageDirection,
      });

      if (!gauge && gauge !== 0) {
        return;
      }

      this.metricsService.incrementGauge("postman_messages_current", 1, {
        status: messageStatus,
        direction: messageDirection,
      });
    } catch (error) {
      this.logger.error("Failed to update metrics:", error);
    }
  }

  private async updateMessageMetricsOnUpdate(event: UpdateEvent<MessageEntity>): Promise<void> {
    try {
      if (!event.entity) {
        return;
      }
      const messageStatus = event.entity.status;
      const previousStatus = event.databaseEntity.status;
      const messageDirection = event.databaseEntity.direction;

      if (messageStatus === MessageStatus.CLAIMED_SUCCESS) {
        this.metricsService.incrementCounter("postman_total_processed_messages", {
          direction: messageDirection,
        });
      }

      const [previousGauge, gauge] = await Promise.all([
        this.metricsService.getGaugeValue("postman_messages_current", {
          status: previousStatus,
          direction: messageDirection,
        }),
        this.metricsService.getGaugeValue("postman_messages_current", {
          status: messageStatus,
          direction: messageDirection,
        }),
      ]);

      if (!previousGauge && previousGauge !== 0) {
        return;
      }

      if (!gauge && gauge !== 0) {
        return;
      }

      if (previousGauge > 0) {
        this.metricsService.decrementGauge("postman_messages_current", 1, {
          status: previousStatus,
          direction: messageDirection,
        });
      }

      this.metricsService.incrementGauge("postman_messages_current", 1, {
        status: messageStatus,
        direction: messageDirection,
      });
    } catch (error) {
      this.logger.error("[MessageStatusSubscriber] Failed to update metrics:", error);
    }
  }

  private async updateMessageMetricsOnRemove(event: RemoveEvent<MessageEntity>): Promise<void> {
    try {
      const messageStatus = event.databaseEntity.status;
      const messageDirection = event.databaseEntity.direction;

      const gauge = await this.metricsService.getGaugeValue("postman_messages_current", {
        status: messageStatus,
        direction: messageDirection,
      });

      if (!gauge && gauge !== 0) {
        return;
      }

      if (gauge > 0) {
        this.metricsService.decrementGauge("postman_messages_current", 1, {
          status: messageStatus,
          direction: messageDirection,
        });
      }
    } catch (error) {
      this.logger.error("Failed to update metrics:", error);
    }
  }
}
