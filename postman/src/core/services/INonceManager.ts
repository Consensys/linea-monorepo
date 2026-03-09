export interface INonceManager {
  acquireNonce(): Promise<number | null>;
  releaseNonce(nonce: number, txHash: string): void;
  reportFailure(nonce: number): void;
}
