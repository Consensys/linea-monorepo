import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor";

export class OssificationCompleteOperationModeProcessor implements IOperationModeProcessor {
  public async poll(): Promise<void> {}
}
