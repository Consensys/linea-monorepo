import { getBlockExtraData } from "@consensys/linea-sdk-viem";
import { BlockTag, type PublicClient } from "viem";

import { ViemProvider } from "../l1/ViemProvider";

import type { ILineaProvider } from "../../../domain/ports/IProvider";
import type { BlockExtraData } from "../../../domain/types/blockchain";

export class ViemLineaProvider extends ViemProvider implements ILineaProvider {
  constructor(publicClient: PublicClient) {
    super(publicClient);
  }

  public async getBlockExtraData(blockNumber: bigint | BlockTag): Promise<BlockExtraData | null> {
    try {
      const params = typeof blockNumber === "string" ? { blockTag: blockNumber } : { blockNumber };

      return await getBlockExtraData(this.publicClient, params);
    } catch {
      return null;
    }
  }
}
