export interface INonceManager {
  getNonce(): Promise<number>;
}
