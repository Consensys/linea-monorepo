export interface IOperationModeProcessor {
  poll(): Promise<void>;
}
