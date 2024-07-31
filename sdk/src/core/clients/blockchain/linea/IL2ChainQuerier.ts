import { IChainQuerier } from "../IChainQuerier";

export type BlockExtraData = {
  version: number;
  fixedCost: number;
  variableCost: number;
  ethGasPrice: number;
};

export interface IL2ChainQuerier<TransactionReceipt, Block, TransactionRequest, TransactionResponse, JsonRpcProvider>
  extends IChainQuerier<TransactionReceipt, Block, TransactionRequest, TransactionResponse, JsonRpcProvider> {
  getBlockExtraData(blockNumber: number | bigint | string): Promise<BlockExtraData | null>;
}
