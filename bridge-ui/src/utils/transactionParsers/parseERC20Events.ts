import { Chain, PublicClient, decodeAbiParameters, getAddress } from "viem";
import log from "loglevel";
import { NetworkTokens, NetworkType } from "@/config/config";
import { ERC20Event, ERC20V2Event } from "@/models";
import { fetchTokenInfo } from "@/services";
import { TransactionHistory } from "@/models/history";
import { findTokenByAddress } from "./helpers";

const parseERC20Events = async (
  events: ERC20Event[],
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
      if (!event.args.token) {
        log.warn("Token args not found");
        return null;
      }

      const tokenAddress = getAddress(event.args.token);
      let token = findTokenByAddress(tokenAddress, storedTokens, networkType);

      // Token list may change, skip old tokens
      if (!token) {
        token = await fetchTokenInfo(tokenAddress, networkType, fromChain);
        if (!token) {
          log.warn("Token not found");
          return null;
        }
      }

      const { timestamp } = await client.getBlock({
        blockNumber: event.blockNumber,
      });

      const [recipient] = decodeAbiParameters([{ type: "bytes32", name: "recipient" }], event.data);

      const logHistory: TransactionHistory = {
        transactionHash: event.transactionHash,
        fromChain,
        toChain,
        tokenAddress,
        token,
        amount: event.args.amount,
        recipient,
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

const parseERC20V2Events = async (
  events: ERC20V2Event[],
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
      if (!event.args.token) {
        log.warn("Token args not found");
        return null;
      }

      const tokenAddress = getAddress(event.args.token);
      let token = findTokenByAddress(tokenAddress, storedTokens, networkType);

      // Token list may change, skip old tokens
      if (!token) {
        token = await fetchTokenInfo(tokenAddress, networkType, fromChain);
        if (!token) {
          log.warn("Token not found");
          return null;
        }
      }

      const { timestamp } = await client.getBlock({
        blockNumber: event.blockNumber,
      });

      const [amount] = decodeAbiParameters([{ type: "uint256", name: "amount" }], event.data);

      const logHistory: TransactionHistory = {
        transactionHash: event.transactionHash,
        fromChain,
        toChain,
        tokenAddress,
        token,
        amount,
        recipient: event.args.recipient,
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

export { parseERC20Events, parseERC20V2Events };
