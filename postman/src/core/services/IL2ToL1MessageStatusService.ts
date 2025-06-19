import { OnChainMessageStatus } from "@consensys/linea-sdk";

export interface IL2ToL1MessageStatusService<Overrides> {
  getMessageStatus(messageHash: string, overrides?: Overrides): Promise<OnChainMessageStatus>;
}

export type L2ToL1MessageStatusServiceConfig = {
  l1LogsFromBlock: number;
  l2LogsFromBlock: number;
};
