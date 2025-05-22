import { Hex } from "viem";

export type Proof = {
  proof: Hex[];
  root: string;
  leafIndex: number;
};
