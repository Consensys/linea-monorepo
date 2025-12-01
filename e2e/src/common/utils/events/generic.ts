import { GetContractEventsParameters, GetContractEventsReturnType, getContractEvents } from "viem/actions";

import { awaitUntil } from "../wait";
import { Abi, Account, BlockNumber, BlockTag, Chain, Client, ContractEventName, Transport } from "viem";

export async function getEvents<
  chain extends Chain | undefined,
  account extends Account | undefined,
  const Tabi extends Abi | readonly unknown[],
  eventName extends ContractEventName<Tabi> | undefined = undefined,
  strict extends boolean | undefined = undefined,
  fromBlock extends BlockNumber | BlockTag | undefined = undefined,
  toBlock extends BlockNumber | BlockTag | undefined = undefined,
>(
  client: Client<Transport, chain, account>,
  params: GetContractEventsParameters<Tabi, eventName, strict, fromBlock, toBlock> & {
    criteria?: (
      events: GetContractEventsReturnType<Tabi, eventName, strict, fromBlock, toBlock>,
    ) => Promise<GetContractEventsReturnType<Tabi, eventName, strict, fromBlock, toBlock>>;
  },
): Promise<GetContractEventsReturnType<Tabi, eventName, strict, fromBlock, toBlock>> {
  const { criteria, ...rest } = params;
  const events = await getContractEvents(client, rest);

  if (criteria) {
    return await criteria(events);
  }

  return events;
}

export async function waitForEvents<
  chain extends Chain | undefined,
  account extends Account | undefined,
  const Tabi extends Abi | readonly unknown[],
  eventName extends ContractEventName<Tabi> | undefined = undefined,
  strict extends boolean | undefined = undefined,
  fromBlock extends BlockNumber | BlockTag | undefined = undefined,
  toBlock extends BlockNumber | BlockTag | undefined = undefined,
>(
  client: Client<Transport, chain, account>,
  params: GetContractEventsParameters<Tabi, eventName, strict, fromBlock, toBlock> & {
    pollingIntervalMs?: number;
    criteria?: (
      events: GetContractEventsReturnType<Tabi, eventName, strict, fromBlock, toBlock>,
    ) => Promise<GetContractEventsReturnType<Tabi, eventName, strict, fromBlock, toBlock>>;
  },
): Promise<GetContractEventsReturnType<Tabi, eventName, strict, fromBlock, toBlock>> {
  return (
    (await awaitUntil(
      async () => await getEvents(client, params),
      (a: GetContractEventsReturnType<Tabi, eventName, strict, fromBlock, toBlock>) => a.length > 0,
      params.pollingIntervalMs ?? 500,
    )) ?? []
  );
}
