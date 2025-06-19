import { OnChainMessageStatus } from "@consensys/linea-sdk";

export interface IL1ToL2MessageStatusService<Overrides> {
  getMessageStatus(messageHash: string, overrides?: Overrides): Promise<OnChainMessageStatus>;
}
