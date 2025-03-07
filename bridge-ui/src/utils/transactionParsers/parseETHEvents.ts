import { PublicClient } from "viem";
import { ETHEvent } from "@/models";
import { TransactionHistory } from "@/models/history";
import { Chain } from "@/types";

const parseETHEvents = async (events: ETHEvent[], client: PublicClient, fromChain: Chain, toChain: Chain) => {
  const history = await Promise.all(
    events.map(async (event) => {
      const { timestamp } = await client.getBlock({
        blockNumber: event.blockNumber,
      });

      const logHistory: TransactionHistory = {
        transactionHash: event.transactionHash,
        fromChain,
        toChain,
        token,
        tokenAddress: null,
        amount: event.args._value,
        recipient: event.args._to,
        pending: true,
        event,
        timestamp,
      };
      return logHistory;
    }),
  );

  return history;
};

export default parseETHEvents;
