import { EventSubscriber, EntitySubscriberInterface, UpdateEvent, RemoveEvent } from "typeorm";
import { Direction } from "@consensys/linea-sdk";
import { MessageEntity } from "../entities/Message.entity";
import { IMetricsService, LineaPostmanMetrics, MetricsOperation } from "../../../../core/metrics/IMetricsService";
import { ILogger } from "../../../../core/utils/logging/ILogger";
import { MessageStatus } from "../../../../core/enums";

@EventSubscriber()
export class SponsorshipFeesSubscriber implements EntitySubscriberInterface<MessageEntity> {
  constructor(
    private readonly metricsService: IMetricsService,
    private readonly logger: ILogger,
  ) {}

  listenTo() {
    return MessageEntity;
  }

  /**
   * We exclusively target the DB event emitted by these lines of code in MessageClaimingPersister:
   
     message.edit({
      status: MessageStatus.CLAIMED_SUCCESS,
      claimTxGasUsed: Number(receipt.gasUsed),
      claimTxGasPrice: BigInt(receipt.gasPrice),
    });
    await this.databaseService.updateMessage(message);
    
   * Hence we don't care about afterTransactionCommit and afterInsert events
   */

  async afterUpdate(event: UpdateEvent<MessageEntity>): Promise<void> {
    if (!event.entity || !event.databaseEntity) return;
    const prevEntityState = event.databaseEntity;
    const newEntityState = event.entity;

    const direction = prevEntityState.direction;
    const newStatus = newEntityState.status;
    const newIsForSponsorship = newEntityState.isForSponsorship;
    const newClaimTxGasUsed = newEntityState?.claimTxGasUsed;
    const newClaimTxGasPrice = newEntityState?.claimTxGasPrice;

    if (newClaimTxGasUsed === undefined || newClaimTxGasPrice === undefined) return;
    if (newStatus !== MessageStatus.CLAIMED_SUCCESS) return;
    if (newIsForSponsorship !== true) return;
    try {
      await this.updateSponsorshipFeeGauges(
        MetricsOperation.INCREMENT,
        direction,
        newClaimTxGasUsed,
        newClaimTxGasPrice,
      );
    } catch (error) {
      this.logger.error("SponsorshipFeesSubscriber.afterUpdate failed:", error);
    }
  }

  async afterRemove(event: RemoveEvent<MessageEntity>): Promise<void> {
    if (event.entity === undefined) return;
    const { status, direction, isForSponsorship, claimTxGasUsed, claimTxGasPrice } = event.entity;
    if (claimTxGasUsed === undefined || claimTxGasPrice === undefined) return;
    if (status !== MessageStatus.CLAIMED_SUCCESS) return;
    if (isForSponsorship !== true) return;
    try {
      await this.updateSponsorshipFeeGauges(MetricsOperation.INCREMENT, direction, claimTxGasUsed, claimTxGasPrice);
    } catch (error) {
      this.logger.error("SponsorshipFeesSubscriber.afterRemove failed:", error);
    }
  }

  private async updateSponsorshipFeeGauges(
    metricsOperation: MetricsOperation,
    direction: Direction,
    claimTxGasUsed: number,
    claimTxGasPrice: bigint,
  ): Promise<void> {
    const [weiGauge, gweiGauge] = await Promise.all([
      this.metricsService.getGaugeValue(LineaPostmanMetrics.SponsorshipFeesWei, { direction }),
      this.metricsService.getGaugeValue(LineaPostmanMetrics.SponsorshipFeesGwei, { direction }),
    ]);
    const { wei, gwei } = this.metricsService.convertTxFeesToWeiAndGwei(
      BigInt(claimTxGasUsed) * BigInt(claimTxGasPrice),
    );
    if (weiGauge !== undefined) {
      switch (metricsOperation) {
        case MetricsOperation.INCREMENT:
          this.metricsService.incrementGauge(LineaPostmanMetrics.SponsorshipFeesWei, { direction }, wei);
          break;
        case MetricsOperation.DECREMENT:
          this.metricsService.decrementGauge(LineaPostmanMetrics.SponsorshipFeesWei, { direction }, wei);
          break;
        default:
          break;
      }
    }
    if (gweiGauge !== undefined) {
      switch (metricsOperation) {
        case MetricsOperation.INCREMENT:
          this.metricsService.incrementGauge(LineaPostmanMetrics.SponsorshipFeesGwei, { direction }, gwei);
          break;
        case MetricsOperation.DECREMENT:
          this.metricsService.decrementGauge(LineaPostmanMetrics.SponsorshipFeesGwei, { direction }, gwei);
          break;
        default:
          break;
      }
    }
  }
}
