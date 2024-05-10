export interface IMessageClaimingProcessor {
  getAndClaimAnchoredMessage(): Promise<void>;
}
