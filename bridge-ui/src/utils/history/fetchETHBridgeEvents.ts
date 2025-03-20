import { Address } from "viem";
import { compareAsc, fromUnixTime, subDays } from "date-fns";
import { getPublicClient } from "@wagmi/core";
import { LineaSDK } from "@consensys/linea-sdk";
import { config as wagmiConfig } from "@/lib/wagmi";
import { defaultTokensConfig } from "@/stores";
import {
  BridgeTransaction,
  BridgeTransactionType,
  Chain,
  ChainLayer,
  Token,
  MessageSentABIEvent,
  MessageSentLogEvent,
} from "@/types";
import { formatOnChainMessageStatus } from "./formatOnChainMessageStatus";

export async function fetchETHBridgeEvents(
  lineaSDK: LineaSDK,
  address: Address,
  fromChain: Chain,
  toChain: Chain,
  tokens: Token[],
): Promise<BridgeTransaction[]> {
  const transactionsMap = new Map<string, BridgeTransaction>();

  const client = getPublicClient(wagmiConfig, {
    chainId: fromChain.id,
  });

  const contract = fromChain.layer === ChainLayer.L1 ? lineaSDK.getL2Contract() : lineaSDK.getL1Contract();

  const messageServiceAddress = fromChain.messageServiceAddress;
  const [ethLogsForSender, ethLogsForRecipient] = await Promise.all([
    client.getLogs({
      event: MessageSentABIEvent,
      fromBlock: "earliest",
      toBlock: "latest",
      address: messageServiceAddress,
      args: {
        _from: address,
      },
    }),
    client.getLogs({
      event: MessageSentABIEvent,
      fromBlock: "earliest",
      toBlock: "latest",
      address: messageServiceAddress,
      args: {
        _to: address,
      },
    }),
  ]);

  const messageSentLogs = [...ethLogsForSender, ...ethLogsForRecipient] as unknown as MessageSentLogEvent[];

  const uniqueLogsMap = new Map<string, (typeof messageSentLogs)[0]>();
  for (const log of messageSentLogs) {
    const uniqueKey = `${log.args._from}-${log.args._to}-${log.transactionHash}`;
    if (!uniqueLogsMap.has(uniqueKey)) {
      uniqueLogsMap.set(uniqueKey, log);
    }
  }

  const currentTimestamp = new Date();

  await Promise.all(
    Array.from(uniqueLogsMap.values()).map(async (log) => {
      const messageHash = log.args._messageHash;

      const block = await client.getBlock({ blockNumber: log.blockNumber, includeTransactions: false });

      if (compareAsc(fromUnixTime(Number(block.timestamp.toString())), subDays(currentTimestamp, 90)) === -1) {
        return;
      }

      const messageStatus = await contract.getMessageStatus(messageHash);

      const token = tokens.find((token) => token.type.includes("eth"));
      const uniqueKey = `${log.args._from}-${log.args._to}-${log.transactionHash}`;
      transactionsMap.set(uniqueKey, {
        type: BridgeTransactionType.ETH,
        status: formatOnChainMessageStatus(messageStatus),
        token: token || defaultTokensConfig.MAINNET[0],
        fromChain,
        toChain,
        timestamp: block.timestamp,
        bridgingTx: log.transactionHash,
        message: {
          from: log.args._from,
          to: log.args._to,
          fee: log.args._fee,
          value: log.args._value,
          nonce: log.args._nonce,
          calldata: log.args._calldata,
          messageHash: log.args._messageHash,
          amountSent: log.args._value,
        },
      });
    }),
  );

  return Array.from(transactionsMap.values());
}
