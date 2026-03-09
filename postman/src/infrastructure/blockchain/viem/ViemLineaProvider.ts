import { getBlockExtraData } from "@consensys/linea-sdk-viem";
import { type BlockTag, type PublicClient } from "viem";

import { ViemProvider } from "./ViemProvider";
import { BlockExtraData, ILineaProvider } from "../../../core/clients/blockchain/linea/ILineaProvider";

export class ViemLineaProvider extends ViemProvider implements ILineaProvider {
  constructor(client: PublicClient) {
    super(client);
  }

  public async getBlockExtraData(blockNumber: number | bigint | string): Promise<BlockExtraData | null> {
    try {
      if (typeof blockNumber === "string") {
        return await getBlockExtraData(this.client, { blockTag: blockNumber as BlockTag });
      }
      return await getBlockExtraData(this.client, { blockNumber: BigInt(blockNumber) });
    } catch {
      return null;
    }
  }
}
