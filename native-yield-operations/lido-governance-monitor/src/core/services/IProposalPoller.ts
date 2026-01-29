export interface IProposalPoller {
  start(): void;
  stop(): void;
  pollOnce(): Promise<void>;
}
