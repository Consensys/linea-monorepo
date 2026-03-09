import { type Hex, type PublicClient } from "viem";

import { Direction } from "../../../core/enums";
import { IMessageDBService } from "../../../core/persistence/IMessageDBService";
import { INonceManager } from "../../../core/services/INonceManager";
import { IPostmanLogger } from "../../../utils/IPostmanLogger";

/**
 * Viem-based nonce manager that wraps the existing inline DB + chain query logic.
 * Preserves behavior from Phase 1 EthersInlineNonceManager while using viem for
 * on-chain nonce lookup instead of ethers Signer.getNonce().
 */
export class InlineNonceManager implements INonceManager {
  private readonly maxNonceDiff: number;

  constructor(
    private readonly databaseService: IMessageDBService,
    private readonly publicClient: PublicClient,
    private readonly signerAddress: string,
    maxNonceDiff: number,
    private readonly direction: Direction,
    private readonly logger: IPostmanLogger,
  ) {
    this.maxNonceDiff = Math.max(maxNonceDiff, 0);
  }

  public async acquireNonce(): Promise<number | null> {
    const [lastTxNonce, onChainNonce] = await Promise.all([
      this.databaseService.getLastClaimTxNonce(this.direction),
      this.publicClient.getTransactionCount({ address: this.signerAddress as Hex }),
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
      this.direction,
      lastTxNonce,
      onChainNonce,
      computedNonce,
    );

    return computedNonce;
  }

  // Phase 1/2: no-ops; in-memory tracking is a future improvement
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public releaseNonce(_nonce: number, _txHash: string): void {}

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public reportFailure(_nonce: number): void {}
}
