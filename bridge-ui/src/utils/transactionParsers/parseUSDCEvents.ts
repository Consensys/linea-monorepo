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
  if (
    networkType !== NetworkType.MAINNET &&
    networkType !== NetworkType.SEPOLIA &&
    networkType !== NetworkType.UNKNOWN
  ) {
    throw new Error("Invalid network type");
  }

  const history = await Promise.all(
    events.map(async (event) => {
      const receipt = await client.getTransactionReceipt({
        hash: event.transactionHash,
      });
      if (!receipt) {
        log.warn(`No receipt found for tx ${event.transactionHash}`);
        return null;
      }

      const tokenAddress = getAddress(receipt.logs[0].address);
      const token = findTokenByAddress(tokenAddress, storedTokens, networkType);

      // Token list may change, skip old tokens
      if (!token) {
        log.warn("Token not found");
        return null;
      }

      const { timestamp } = await client.getBlock({
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
        timestamp,
      };

      return logHistory;
    }),
  );

  const newHistory: TransactionHistory[] = [];

  for (const event of history) {
    if (event) {
      newHistory.push(event);
    }
  }

  return newHistory;
};

export default parseUSDCEvents;
