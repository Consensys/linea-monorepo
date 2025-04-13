import { EntityManager } from "typeorm";
import { Direction } from "@consensys/linea-sdk";
import { MetricService } from "./MetricService";
import { MessageEntity } from "../../persistence/entities/Message.entity";
import { MessageStatus } from "../../../../core/enums";

export class MessageMetricsService extends MetricService {
  constructor(private readonly entityManager: EntityManager) {
    super();
    this.createCounter("postman_total_processed_messages", "Total number of messages by status", ["direction"]);
    this.createGauge("postman_messages_current", "Current number of messages by status", ["status", "direction"]);
  }

  public async initialize(): Promise<void> {
    const totalNumberOfProcessedMessages = await this.getTotalProcessedMessages();
    if (totalNumberOfProcessedMessages) {
      this.incrementCounter(
        "postman_total_processed_messages",
        {
          direction: Direction.L1_TO_L2,
        },
        totalNumberOfProcessedMessages,
      );
    }

    const fullResult = await this.getMessagesCountByStatus();
    this.updateGauges(fullResult);
  }

  private async getTotalProcessedMessages(): Promise<number | null> {
    return this.entityManager.maximum(MessageEntity, "messageNonce", {
      direction: Direction.L1_TO_L2,
    });
  }

  private async getMessagesCountByStatus(): Promise<{ status: string; count: number }[]> {
    const totalNumberOfL1L2MessagesByStatus = await this.entityManager
      .createQueryBuilder(MessageEntity, "message")
      .select("message.status", "status")
      .addSelect("COUNT(message.id)", "count")
      .groupBy("message.status")
      .getRawMany();

    const resultMap = new Map(totalNumberOfL1L2MessagesByStatus.map((r) => [r.status, Number(r.count)]));

    return Object.values(MessageStatus).map((status) => ({
      status,
      count: resultMap.get(status) || 0,
    }));
  }

  private updateGauges(fullResult: { status: string; count: number }[]): void {
    for (const { status, count } of fullResult) {
      this.incrementGauge("postman_messages_current", count, {
        status,
        direction: Direction.L1_TO_L2,
      });
    }
  }
}
