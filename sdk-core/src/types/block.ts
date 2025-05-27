export type BlockNumber<unit = bigint> = unit;

export type BlockTag = "latest" | "earliest" | "pending" | "safe" | "finalized";

export type BlockExtraData = {
  version: number;
  fixedCost: number;
  variableCost: number;
  ethGasPrice: number;
};
