import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor";

export class OssificationPendingOperationModeProcessor implements IOperationModeProcessor {
  public async poll(): Promise<void> {}
}
