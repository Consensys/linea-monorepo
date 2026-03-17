import { ILogger } from "@consensys/linea-shared-utils";

import { ITransactionCountProvider } from "../../../../core/clients/blockchain/IProvider";
import { INonceManager } from "../../../../core/services/INonceManager";
import { Address } from "../../../../core/types";

export class InMemoryNonceManager implements INonceManager {
  private nextNonce = 0;
  private readonly reusable: number[] = [];
  private readonly pending = new Set<number>();
  private readonly maxNonceDiff: number;

  private locked = false;
  private readonly waitQueue: Array<(value: void) => void> = [];

  constructor(
    private readonly provider: ITransactionCountProvider,
    private readonly signerAddress: Address,
    maxNonceDiff: number,
    private readonly logger: ILogger,
  ) {
    this.maxNonceDiff = Math.max(maxNonceDiff, 0);
  }

  public async initialize(): Promise<void> {
    const onChainNonce = await this.provider.getTransactionCount(this.signerAddress, "pending");
    this.nextNonce = onChainNonce;
    this.logger.info("NonceManager initialized.", { startNonce: onChainNonce });
  }

  public async acquireNonce(): Promise<number> {
    await this.lock();
    try {
      const onChainNonce = await this.provider.getTransactionCount(this.signerAddress, "pending");
      const drift = this.nextNonce - onChainNonce;

      if (drift > this.maxNonceDiff && this.reusable.length === 0) {
        this.logger.warn("Nonce drift exceeds limit, resynchronizing with on-chain nonce.", {
          nextNonce: this.nextNonce,
          onChainNonce,
          maxNonceDiff: this.maxNonceDiff,
          pendingCount: this.pending.size,
        });
        this.nextNonce = onChainNonce;
        this.pending.clear();
      }

      let nonce: number;
      if (this.reusable.length > 0) {
        nonce = this.reusable.shift()!;
      } else {
        nonce = this.nextNonce++;
      }

      this.pending.add(nonce);

      this.logger.debug("Nonce acquired.", {
        nonce,
        nextNonce: this.nextNonce,
        pendingCount: this.pending.size,
        reusableCount: this.reusable.length,
      });

      return nonce;
    } finally {
      this.unlock();
    }
  }

  public commitNonce(nonce: number): void {
    this.pending.delete(nonce);
    this.logger.debug("Nonce committed.", { nonce, pendingCount: this.pending.size });
  }

  public rollbackNonce(nonce: number): void {
    this.pending.delete(nonce);
    const idx = this.reusable.findIndex((n) => n > nonce);
    if (idx === -1) {
      this.reusable.push(nonce);
    } else {
      this.reusable.splice(idx, 0, nonce);
    }
    this.logger.debug("Nonce rolled back.", { nonce, reusableCount: this.reusable.length });
  }

  private async lock(): Promise<void> {
    if (!this.locked) {
      this.locked = true;
      return;
    }
    return new Promise<void>((resolve) => {
      this.waitQueue.push(resolve);
    });
  }

  private unlock(): void {
    const next = this.waitQueue.shift();
    if (next) {
      next();
    } else {
      this.locked = false;
    }
  }
}
