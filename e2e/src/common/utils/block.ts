import { Client, Transport, GetBlockParameters, GetBlockReturnType, Chain, Account } from "viem";
import { getBlockNumber, getBlock } from "viem/actions";
import { awaitUntil } from "./wait";

export async function pollForBlockNumber(
  client: Client,
  expectedBlockNumber: bigint,
  pollingIntervalMs: number = 500,
  timeoutMs: number = 2 * 60 * 1000,
): Promise<boolean> {
  return (
    (await awaitUntil(
      async () => await getBlockNumber(client),
      (a: bigint) => a >= expectedBlockNumber,
      pollingIntervalMs,
      timeoutMs,
    )) != null
  );
}

export async function getBlockByNumberOrBlockTag<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<Transport, chain, account>,
  params: GetBlockParameters,
): Promise<GetBlockReturnType | null> {
  try {
    const blockNumber = await getBlock(client, params);
    return blockNumber;
  } catch (error) {
    return null;
  }
}
