import { Chain, PublicClient } from "viem";
import log from "loglevel";
import { NetworkTokens, NetworkType } from "@/config";
import { ETHEvent } from "@/models";
import { TransactionHistory } from "@/models/history";
import { findETHToken } from "./helpers";

// Import other dependencies and methods...

const parseETHEvents = async (
  events: ETHEvent[],
  client: PublicClient,
  fromChain: Chain,
  toChain: Chain,
  storedTokens: NetworkTokens,
  networkType: NetworkType,
) => {
  const history: TransactionHistory[] = [];

  if (
    networkType !== NetworkType.MAINNET &&
    networkType !== NetworkType.SEPOLIA &&
    networkType !== NetworkType.UNKNOWN
  ) {
    throw new Error("Invalid network type");
  }

  for (const event of events) {
    const token = findETHToken(storedTokens, networkType);

    // Token list may change, skip old tokens
    if (!token) {
      log.warn("Token not found");
      continue;
    }

    const blockInfo = await client.getBlock({
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
      timestamp: blockInfo.timestamp,
    };
    history.push(logHistory);
  }
  return history;
};

export default parseETHEvents;
