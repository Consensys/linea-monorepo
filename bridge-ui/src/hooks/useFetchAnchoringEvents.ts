import { getPublicClient, getBlockNumber } from "@wagmi/core";
import { Address, Chain } from "viem";
import { eventMessageHashes } from "@/utils/transactionParsers";
import { wagmiConfig } from "@/config";

const BLOCK_LIMIT = BigInt(5);

const useFetchAnchoringEvents = () => {
  const fetchAnchoringMessageHashes = async (
    chain: Chain,
    messageServiceAddress: Address | null,
  ): Promise<string[]> => {
    if (!messageServiceAddress) {
      return [];
    }

    const client = getPublicClient(wagmiConfig, {
      chainId: chain.id,
    });

    if (!client) {
      return [];
    }

    const toBlock = await getBlockNumber(wagmiConfig, {
      chainId: chain.id,
    });
    const fromBlock = toBlock - BLOCK_LIMIT;

    const logs = await client.getLogs({
      event: eventMessageHashes,
      fromBlock,
      toBlock,
      address: messageServiceAddress,
    });

    const messageHashes: string[] = logs.flatMap((log) => log.args.messageHashes || []);

    return messageHashes;
  };

  return { fetchAnchoringMessageHashes };
};

export default useFetchAnchoringEvents;
