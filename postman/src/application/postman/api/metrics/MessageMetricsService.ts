import { EntityManager } from "typeorm";
import { MetricsService } from "./MetricsService";
import { MessageEntity } from "../../persistence/entities/Message.entity";
import {
  LineaPostmanMetrics,
  MessagesMetricsAttributes,
  SponsorshipFeesMetricsAttributes,
} from "../../../../core/metrics/IMetricsService";
import { Direction } from "@consensys/linea-sdk";
import { MessageStatus } from "../../../../core/enums";

export class MessageMetricsService extends MetricsService {
  // FOR LATER - How to enable search by time for these metrics
  constructor(private readonly entityManager: EntityManager) {
    super();
    this.createGauge(
      LineaPostmanMetrics.Messages,
      "Current number of messages by status, direction and sponsorship status",
      ["status", "direction", "isForSponsorship"],
    );
    this.createGauge(
      LineaPostmanMetrics.SponsorshipFeesWei,
      "Current wei component of tx fees paid for sponsored messages by direction",
      ["direction"],
    );
    this.createGauge(
      LineaPostmanMetrics.SponsorshipFeesGwei,
      "Current gwei component of tx fees paid for sponsored messages by direction",
      ["direction"],
    );
  }

  public async initialize(): Promise<void> {
    this.initializeMessagesGauges();
    this.initializeSponsorshipFeesGauges();
  }

  private async initializeSponsorshipFeesGauges(): Promise<void> {
    const totalNumberOfMessagesByAttributeGroups = await this.entityManager
      .createQueryBuilder(MessageEntity, "message")
      .select("message.direction", "direction")
      .addSelect("SUM((message.claim_tx_gas_used::bigint) * message.claim_tx_gas_price)", "totalTxFees")
      // Only include CLAIMED_SUCCESS messages which were sponsored
      .where("message.status = :status", { status: MessageStatus.CLAIMED_SUCCESS })
      .andWhere("message.is_for_sponsorship = true")
      .groupBy("message.direction")
      .getRawMany();

    // JSON.stringify(SponsorshipFeesMetricsAttributes) => Count
    const weiResultMap = new Map<string, number>();
    const gweiResultMap = new Map<string, number>();

    totalNumberOfMessagesByAttributeGroups.forEach((r) => {
      const sponsorshipFeeMetricAttributesWei: SponsorshipFeesMetricsAttributes = {
        direction: r.direction,
      };
      const sponsorshipFeeMetricAttributesGwei: SponsorshipFeesMetricsAttributes = {
        direction: r.direction,
      };
      const { wei, gwei } = this.convertTxFeesToWeiAndGwei(r.totalTxFees);
      const resultMapKeyWei = JSON.stringify(sponsorshipFeeMetricAttributesWei);
      const resultMapKeyGwei = JSON.stringify(sponsorshipFeeMetricAttributesGwei);
      weiResultMap.set(resultMapKeyWei, wei);
      gweiResultMap.set(resultMapKeyGwei, gwei);
    });

    // Note that we must initialize every attribute combination, or 'incrementGauge' and 'decrementGauge' will not work later on.
    for (const direction of Object.values(Direction)) {
      const attributes: SponsorshipFeesMetricsAttributes = {
        direction,
      };
      const attributesKey = JSON.stringify(attributes);
      this.incrementGauge(
        LineaPostmanMetrics.SponsorshipFeesWei,
        {
          direction: direction,
        },
        weiResultMap.get(attributesKey) ?? 0,
      );
      this.incrementGauge(
        LineaPostmanMetrics.SponsorshipFeesGwei,
        {
          direction: direction,
        },
        gweiResultMap.get(attributesKey) ?? 0,
      );
    }
  }

  // TO CONSIDER IN LATER TICKET - Some combinations of (status,direction,isForSponsorship) should not happen. Should we still create the metric for these combinations?
  private async initializeMessagesGauges(): Promise<void> {
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
          this.incrementGauge(
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
}
