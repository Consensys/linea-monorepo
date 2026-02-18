import { IMetricsService } from "@consensys/linea-shared-utils";
import { EntityManager } from "typeorm";

import { IMessageMetricsUpdater, LineaPostmanMetrics, MessagesMetricsAttributes } from "../../domain/ports/IMetrics";
import { Direction, MessageStatus } from "../../domain/types/enums";
import { MessageEntity } from "../persistence/entities/MessageEntity";

export class MessageMetricsUpdater implements IMessageMetricsUpdater {
  constructor(
    private readonly entityManager: EntityManager,
    private readonly metricsService: IMetricsService<LineaPostmanMetrics>,
  ) {
    this.metricsService.createGauge(
      LineaPostmanMetrics.Messages,
      "Current number of messages by status and direction",
      ["status", "direction"],
    );
  }

  public async initialize(): Promise<void> {
    await this.initializeMessagesGauges();
  }

  private async initializeMessagesGauges(): Promise<void> {
    const totalNumberOfMessagesByAttributeGroups = await this.entityManager
      .createQueryBuilder(MessageEntity, "message")
      .select("message.status", "status")
      .addSelect("message.direction", "direction")
      .addSelect("COUNT(message.id)", "count")
      .groupBy("message.status")
      .addGroupBy("message.direction")
      .getRawMany();

    const resultMap = new Map<string, number>();

    totalNumberOfMessagesByAttributeGroups.forEach((r: { status: string; direction: string; count: string }) => {
      const messageMetricAttributes: MessagesMetricsAttributes = {
        status: r.status as MessageStatus,
        direction: r.direction as Direction,
      };
      const resultMapKey = JSON.stringify(messageMetricAttributes);
      resultMap.set(resultMapKey, parseInt(r.count));
    });

    for (const status of Object.values(MessageStatus)) {
      for (const direction of Object.values(Direction)) {
        const attributes: MessagesMetricsAttributes = {
          status,
          direction,
        };
        const attributesKey = JSON.stringify(attributes);
        this.metricsService.incrementGauge(
          LineaPostmanMetrics.Messages,
          {
            status: attributes.status,
            direction: attributes.direction,
          },
          resultMap.get(attributesKey) ?? 0,
        );
      }
    }
  }

  public async getMessageCount(messageAttributes: MessagesMetricsAttributes): Promise<number | undefined> {
    const { status, direction } = messageAttributes;
    return await this.metricsService.getGaugeValue(LineaPostmanMetrics.Messages, {
      status,
      direction,
    });
  }

  public async incrementMessageCount(messageAttributes: MessagesMetricsAttributes, value: number = 1): Promise<void> {
    const { status, direction } = messageAttributes;
    this.metricsService.incrementGauge(
      LineaPostmanMetrics.Messages,
      {
        status,
        direction,
      },
      value,
    );
  }

  public async decrementMessageCount(messageAttributes: MessagesMetricsAttributes, value: number = 1): Promise<void> {
    const { status, direction } = messageAttributes;
    this.metricsService.decrementGauge(
      LineaPostmanMetrics.Messages,
      {
        status,
        direction,
      },
      value,
    );
  }
}
