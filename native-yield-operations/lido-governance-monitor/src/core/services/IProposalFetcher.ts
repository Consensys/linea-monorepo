export interface IProposalFetcher {
  pollOnce(): Promise<void>;
}
