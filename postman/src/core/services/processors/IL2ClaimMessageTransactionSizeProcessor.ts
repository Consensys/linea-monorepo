import { Direction } from "../../enums";

import type { Address } from "../../types/hex";

export interface IL2ClaimMessageTransactionSizeProcessor {
  process(): Promise<void>;
}

export type L2ClaimMessageTransactionSizeProcessorConfig = {
  direction: Direction;
  originContractAddress: Address;
};
