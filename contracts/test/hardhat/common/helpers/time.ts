import { networkHelpers } from "../connection.js";
const { time } = networkHelpers;

export const getLastBlockTimestamp = async (): Promise<bigint> => {
  return BigInt(await time.latest());
};

export const setFutureTimestampForNextBlock = async (secondsInTheFuture: number | bigint = 1): Promise<bigint> => {
  const lastBlockTimestamp: number = await time.latest();
  const futureTimestamp: bigint = BigInt(lastBlockTimestamp) + BigInt(secondsInTheFuture);
  await time.setNextBlockTimestamp(futureTimestamp);
  return futureTimestamp;
};
