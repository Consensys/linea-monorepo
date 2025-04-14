import { EntityManager } from "typeorm";
import { Direction } from "@consensys/linea-sdk";
import { MetricsService } from "./MetricsService";
import { MessageEntity } from "../../persistence/entities/Message.entity";
import { MessageStatus } from "../../../../core/enums";

export class MessageMetricsService extends MetricsService {
  constructor(private readonly entityManager: EntityManager) {
    super();
    this.createGauge("postman_messages_current", "Current number of messages by status", ["status", "direction"]);
  }

  public async initialize(): Promise<void> {
    const fullResult = await this.getMessagesCountByStatus();
    this.updateGauges(fullResult);
  }

  private async getMessagesCountByStatus(): Promise<{ status: MessageStatus; direction: Direction; count: number }[]> {
    const totalNumberOfMessagesByStatusAndDirection = await this.entityManager
      .createQueryBuilder(MessageEntity, "message")
      .select("message.status", "status")
      .addSelect("message.direction", "direction")
      .addSelect("COUNT(message.id)", "count")
      .groupBy("message.status")
      .addGroupBy("message.direction")
      .getRawMany();

    const resultMap = new Map<string, Map<string, number>>();

    totalNumberOfMessagesByStatusAndDirection.forEach((r) => {
      if (!resultMap.has(r.status)) {
        resultMap.set(r.status, new Map());
      }
      resultMap.get(r.status)!.set(r.direction, Number(r.count));
    });

    const results: { status: MessageStatus; direction: Direction; count: number }[] = [];

    for (const status of Object.values(MessageStatus)) {
      for (const direction of Object.values(Direction)) {
        results.push({
          status,
          direction,
          count: resultMap.get(status)?.get(direction) || 0,
        });
      }
    }

    return results;
  }

  private updateGauges(fullResult: { status: MessageStatus; direction: Direction; count: number }[]): void {
    for (const { status, count, direction } of fullResult) {
      this.incrementGauge("postman_messages_current", count, {
        status,
        direction,
      });
    }
  }
}
