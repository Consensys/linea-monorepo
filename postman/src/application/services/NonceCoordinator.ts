import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";
import type { INonceManager } from "../../domain/ports/INonceManager";
import type { Direction } from "../../domain/types/enums";

export class NonceCoordinator {
  private readonly maxNonceDiff: number;

  constructor(
    private readonly repository: IMessageRepository,
    private readonly nonceManager: INonceManager,
    maxNonceDiff: number,
    private readonly logger: ILogger,
  ) {
    this.maxNonceDiff = Math.max(maxNonceDiff, 0);
  }

  public async getNextNonce(direction: Direction): Promise<number | null> {
    const [lastTxNonce, onChainNonce] = await Promise.all([
      this.repository.getLastClaimTxNonce(direction),
      this.nonceManager.getNonce(),
    ]);

    if (lastTxNonce === null) {
      return onChainNonce;
    }

    if (lastTxNonce - onChainNonce > this.maxNonceDiff) {
      this.logger.warn(
        "Last recorded nonce in db is higher than the latest nonce from blockchain and exceeds the diff limit, paused the claim message process now: nonceInDb=%s nonceOnChain=%s maxAllowedNonceDiff=%s",
        lastTxNonce,
        onChainNonce,
        this.maxNonceDiff,
      );
      return null;
    }

    const computedNonce = Math.max(onChainNonce, lastTxNonce + 1);

    this.logger.debug(
      "Nonce computation: direction=%s lastTxNonce=%s onChainNonce=%s computedNonce=%s",
      direction,
      lastTxNonce,
      onChainNonce,
      computedNonce,
    );

    return computedNonce;
  }
}
