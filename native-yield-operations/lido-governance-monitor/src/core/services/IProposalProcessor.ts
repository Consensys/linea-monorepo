export interface IProposalProcessor {
  start(): void;
  stop(): void;
  processOnce(): Promise<void>;
}
