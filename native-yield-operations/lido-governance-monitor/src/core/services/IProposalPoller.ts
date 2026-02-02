export interface IProposalPoller {
  pollOnce(): Promise<void>;
}
