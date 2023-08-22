import { BigNumber, BytesLike } from "ethers";

export type SendMessageArgs = {
  to: string;
  fee: BigNumber;
  calldata: BytesLike;
};
