export interface IMessageAnchoringProcessor {
  getAndUpdateAnchoredMessageStatus(): Promise<void>;
}
