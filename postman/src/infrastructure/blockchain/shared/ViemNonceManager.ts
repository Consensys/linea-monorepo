import { type PublicClient, type Address } from "viem";

import type { INonceManager } from "../../../domain/ports/INonceManager";

export class ViemNonceManager implements INonceManager {
  constructor(
    private readonly publicClient: PublicClient,
    private readonly address: Address,
  ) {}

  public async getNonce(): Promise<number> {
    return this.publicClient.getTransactionCount({
      address: this.address,
      blockTag: "pending",
    });
  }
}
