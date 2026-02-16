import { getBlockExtraData } from "@consensys/linea-sdk-viem";
import { type PublicClient, type Client } from "viem";

import { ViemProvider } from "./ViemProvider";

import type { ILineaProvider } from "../../../domain/ports/IProvider";
import type { BlockExtraData } from "../../../domain/types";

export class ViemLineaProvider extends ViemProvider implements ILineaProvider {
  constructor(publicClient: PublicClient) {
    super(publicClient);
  }

  public async getBlockExtraData(blockNumber: number | bigint | string): Promise<BlockExtraData | null> {
    try {
      const params = blockNumber === "latest" ? { blockTag: "latest" as const } : { blockNumber: BigInt(blockNumber) };

      return await getBlockExtraData(this.publicClient as Client, params);
    } catch {
      return null;
    }
  }
}
