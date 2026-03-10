import { getBlockExtraData } from "@consensys/linea-sdk-viem";
import { BlockNumber, type BlockTag, type PublicClient } from "viem";

import { ViemProvider } from "./ViemProvider";
import { BlockExtraData, ILineaProvider } from "../../../../core/clients/blockchain/linea/ILineaProvider";

export class ViemLineaProvider extends ViemProvider implements ILineaProvider {
  constructor(client: PublicClient) {
    super(client);
  }

  public async getBlockExtraData(blockNumber: BlockNumber | BlockTag): Promise<BlockExtraData | null> {
    try {
      if (typeof blockNumber === "string") {
        return await getBlockExtraData(this.client, { blockTag: blockNumber });
      }
      return await getBlockExtraData(this.client, { blockNumber: BigInt(blockNumber) });
    } catch {
      return null;
    }
  }
}
