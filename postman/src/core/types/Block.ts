import type { Hash } from "./hex";

export type Block = {
  number: number;
  timestamp: number;
  hash: Hash;
};
