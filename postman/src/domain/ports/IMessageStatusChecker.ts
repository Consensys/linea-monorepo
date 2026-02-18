import { OnChainMessageStatus } from "../types/enums";

export interface IMessageStatusChecker {
  getMessageStatus(params: { messageHash: string; messageBlockNumber?: number }): Promise<OnChainMessageStatus>;
}
