import { EntityManager } from "typeorm";
import { MetricsService } from "./MetricsService";
import { MessageEntity } from "../../persistence/entities/Message.entity";
import {
  LineaPostmanMetrics,
  MessagesMetricsAttributes,
  MessagesMetricsAttributesWithCount,
} from "../../../../core/metrics/IMetricsService";
export class MessageMetricsService extends MetricsService {
  constructor(private readonly entityManager: EntityManager) {
    super();
    this.createGauge(
      LineaPostmanMetrics.Messages,
      "Current number of messages by status, direction and sponsorship status",
      ["status", "direction", "isForSponsorship"],
    );
  }

  public async initialize(): Promise<void> {
    const messagesByAttribute = await this.getMessagesCountFromDatabase();
    this.initializeGaugeValues(messagesByAttribute);
  }

  private async getMessagesCountFromDatabase(): Promise<MessagesMetricsAttributesWithCount[]> {
    const totalNumberOfMessagesByAttributeGroups = await this.entityManager
      .createQueryBuilder(MessageEntity, "message")
      .select("message.status", "status")
      .addSelect("message.direction", "direction")
      .addSelect("message.is_for_sponsorship", "isForSponsorship")
      .addSelect("COUNT(message.id)", "count")
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
      resultMap.set(resultMapKey, r.count);
    });

    const results: MessagesMetricsAttributesWithCount[] = [];

    resultMap.forEach((count, resultMapKey) => {
      const attributes: MessagesMetricsAttributes = JSON.parse(resultMapKey);
      results.push({
        attributes,
        count,
      });
    });

    return results;
  }

  private initializeGaugeValues(messagesByAttribute: MessagesMetricsAttributesWithCount[]): void {
    for (const { attributes, count } of messagesByAttribute) {
      this.incrementGauge(
        LineaPostmanMetrics.Messages,
        {
          status: attributes.status,
          direction: attributes.direction,
          isForSponsorship: String(attributes.isForSponsorship),
        },
        count,
      );
    }
  }
}
