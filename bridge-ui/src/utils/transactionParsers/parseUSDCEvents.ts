import { Chain, PublicClient, getAddress } from "viem";
import log from "loglevel";
import { NetworkTokens, NetworkType } from "@/config";
import { USDCEvent } from "@/models";
import { TransactionHistory } from "@/models/history";
import { findTokenByAddress } from "./helpers";

const parseUSDCEvents = async (
  events: USDCEvent[],
  client: PublicClient,
  fromChain: Chain,
  toChain: Chain,
  storedTokens: NetworkTokens,
  networkType: NetworkType,
) => {
  const newHistory: TransactionHistory[] = [];

  if (
    networkType !== NetworkType.MAINNET &&
    networkType !== NetworkType.SEPOLIA &&
    networkType !== NetworkType.UNKNOWN
  ) {
    throw new Error("Invalid network type");
  }

  for (const event of events) {
    const receipt = await client.getTransactionReceipt({
      hash: event.transactionHash,
    });
    if (!receipt) {
      log.warn(`No receipt found for tx ${event.transactionHash}`);
    }

    const tokenAddress = getAddress(receipt.logs[0].address);
    const token = findTokenByAddress(tokenAddress, storedTokens, networkType);

    // Token list may change, skip old tokens
    if (!token) {
      log.warn("Token not found");
      continue;
    }

    // Get block timestamp
    const blockInfo = await client.getBlock({
      blockNumber: receipt.blockNumber,
    });

    const logHistory: TransactionHistory = {
      transactionHash: event.transactionHash,
      fromChain,
      toChain,
      tokenAddress,
      token,
      amount: event.args.amount,
      recipient: event.args.to,
      pending: true,
      event,
      timestamp: blockInfo.timestamp,
    };
    newHistory.push(logHistory);
  }

  return newHistory;
};

export default parseUSDCEvents;
