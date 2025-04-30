import { Direction } from "@consensys/linea-sdk";

export interface ISponsorshipMetricsUpdater {
  getSponsorshipFeePaid(direction: Direction): Promise<bigint>;
  incrementSponsorshipFeePaid(txFee: bigint, direction: Direction): Promise<void>;
}
