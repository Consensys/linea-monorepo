export interface IMessageClaimingPersister {
  updateAndPersistPendingMessage(): Promise<void>;
}
