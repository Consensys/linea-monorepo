import { EntityManager } from "typeorm";
import { MessageEntity } from "../../persistence/entities/Message.entity";
import {
  IMetricsService,
  IMessageMetricsUpdater,
  LineaPostmanMetrics,
  MessagesMetricsAttributes,
} from "../../../../core/metrics";
import { Direction } from "@consensys/linea-sdk";
import { MessageStatus } from "../../../../core/enums";

export class MessageMetricsUpdater implements IMessageMetricsUpdater {
  constructor(
    private readonly entityManager: EntityManager,
    private readonly metricsService: IMetricsService,
  ) {
    this.metricsService.createGauge(
      LineaPostmanMetrics.Messages,
      "Current number of messages by status, direction and sponsorship status",
      ["status", "direction", "isForSponsorship"],
    );
  }

  public async initialize(): Promise<void> {
    this.initializeMessagesGauges();
  }

  // TO CONSIDER IN LATER TICKET - Some combinations of (status,direction,isForSponsorship) should not happen. Should we still create the metric for these combinations?
  // TODO - Consider isForSponsorship attribute
  private async initializeMessagesGauges(): Promise<void> {
    const totalNumberOfMessagesByAttributeGroups = await this.entityManager
      .createQueryBuilder(MessageEntity, "message")
      .select("message.status", "status")
      .addSelect("message.direction", "direction")
      .addSelect("message.is_for_sponsorship", "isForSponsorship")
      .addSelect("COUNT(message.id)", "count") // Actually a string type
      .groupBy("message.status")
      .addGroupBy("message.direction")
      .addGroupBy("message.is_for_sponsorship")
      .getRawMany();

    // JSON.stringify(MessagesMetricsAttributes) => Count
    const resultMap = new Map<string, number>();

    totalNumberOfMessagesByAttributeGroups.forEach((r) => {
      const messageMetricAttributes: MessagesMetricsAttributes = {
        status: r.status,
        direction: r.direction,
        isForSponsorship: r.isForSponsorship,
      };
      const resultMapKey = JSON.stringify(messageMetricAttributes);
      resultMap.set(resultMapKey, parseInt(r.count));
    });

    // Note that we must initialize every attribute combination, or 'incrementGauge' and 'decrementGauge' will not work later on.
    for (const status of Object.values(MessageStatus)) {
      for (const direction of Object.values(Direction)) {
        for (const isForSponsorship of [true, false]) {
          const attributes: MessagesMetricsAttributes = {
            status,
            direction,
            isForSponsorship,
          };
          const attributesKey = JSON.stringify(attributes);
          this.metricsService.incrementGauge(
            LineaPostmanMetrics.Messages,
            {
              status: attributes.status,
              direction: attributes.direction,
              isForSponsorship: String(attributes.isForSponsorship),
            },
            resultMap.get(attributesKey) ?? 0,
          );
        }
      }
    }
  }

  public async getMessageCount(messageAttributes: MessagesMetricsAttributes): Promise<number | undefined> {
    const { status, direction, isForSponsorship } = messageAttributes;
    return await this.metricsService.getGaugeValue(LineaPostmanMetrics.Messages, {
      status,
      direction,
      isForSponsorship: String(isForSponsorship),
    });
  }

  public async incrementMessageCount(messageAttributes: MessagesMetricsAttributes, value: number = 1) {
    const { status, direction, isForSponsorship } = messageAttributes;
    await this.metricsService.incrementGauge(
      LineaPostmanMetrics.Messages,
      {
        status,
        direction,
        isForSponsorship: String(isForSponsorship),
      },
      value,
    );
  }

  public async decrementMessageCount(messageAttributes: MessagesMetricsAttributes, value: number = 1) {
    const { status, direction, isForSponsorship } = messageAttributes;
    await this.metricsService.decrementGauge(
      LineaPostmanMetrics.Messages,
      {
        status,
        direction,
        isForSponsorship: String(isForSponsorship),
      },
      value,
    );
  }
}
