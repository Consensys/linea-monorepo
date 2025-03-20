import { Address, decodeAbiParameters } from "viem";
import { compareAsc, fromUnixTime, subDays } from "date-fns";
import { getPublicClient } from "@wagmi/core";
import { LineaSDK, OnChainMessageStatus } from "@consensys/linea-sdk";
import { config as wagmiConfig } from "@/lib/wagmi";
import { config } from "@/config";
import { defaultTokensConfig } from "@/stores";
import {
  BridgeTransaction,
  BridgeTransactionType,
  Chain,
  ChainLayer,
  Token,
  TransactionStatus,
  BridgingInitiatedV2LogEvent,
  BridgingInitiatedV2ABIEvent,
  MessageSentABIEvent,
  CCTPDepositForBurnAbiEvent,
  MessageSentLogEvent,
} from "@/types";
import { isCctp, isCCTPNonceUsed, getCCTPTransactionStatus, getCCTPMessageByTxHash } from "@/utils";
import { DepositForBurnLogEvent } from "@/types/events";

type TransactionHistoryParams = {
  lineaSDK: LineaSDK;
  fromChain: Chain;
  toChain: Chain;
  address: Address;
  tokens: Token[];
};

export async function fetchTransactionsHistory({
  lineaSDK,
  fromChain,
  toChain,
  address,
  tokens,
}: TransactionHistoryParams): Promise<BridgeTransaction[]> {
  const events = await Promise.all([
    fetchBridgeEvents(lineaSDK, fromChain, toChain, address, tokens),
    fetchBridgeEvents(lineaSDK, toChain, fromChain, address, tokens),
  ]);
  return events.flat().sort((a, b) => Number(b.timestamp.toString()) - Number(a.timestamp.toString()));
}

// TODO - Memoize events
async function fetchBridgeEvents(
  lineaSDK: LineaSDK,
  fromChain: Chain,
  toChain: Chain,
  address: Address,
  tokens: Token[],
): Promise<BridgeTransaction[]> {
  const [ethEvents, erc20Events, cctpEvents] = await Promise.all([
    fetchETHBridgeEvents(lineaSDK, address, fromChain, toChain, tokens),
    fetchERC20BridgeEvents(lineaSDK, address, fromChain, toChain, tokens),
    // Feature toggle for CCTP, will filter out USDC transactions if isCCTPEnabled == false
    config.isCCTPEnabled ? fetchCCTPBridgeEvents(address, fromChain, toChain, tokens) : [],
  ]);

  return [...ethEvents, ...erc20Events, ...cctpEvents];
}

async function fetchETHBridgeEvents(
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
        status: formatStatus(messageStatus),
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

async function fetchERC20BridgeEvents(
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

  const [originContract, destinationContract] =
    fromChain.layer === ChainLayer.L1
      ? [lineaSDK.getL1Contract(), lineaSDK.getL2Contract()]
      : [lineaSDK.getL2Contract(), lineaSDK.getL1Contract()];

  const tokenBridgeAddress = fromChain.tokenBridgeAddress;
  const [erc20LogsForSender, erc20LogsForRecipient] = await Promise.all([
    client.getLogs({
      event: BridgingInitiatedV2ABIEvent,
      fromBlock: "earliest",
      toBlock: "latest",
      address: tokenBridgeAddress,
      args: {
        sender: address,
      },
    }),
    client.getLogs({
      event: BridgingInitiatedV2ABIEvent,
      fromBlock: "earliest",
      toBlock: "latest",
      address: tokenBridgeAddress,
      args: {
        recipient: address,
      },
    }),
  ]);

  const erc20Logs = [...erc20LogsForSender, ...erc20LogsForRecipient] as unknown as BridgingInitiatedV2LogEvent[];

  const uniqueLogsMap = new Map<string, (typeof erc20Logs)[0]>();
  for (const log of erc20Logs) {
    const transactionHash = log.transactionHash;
    if (!uniqueLogsMap.has(transactionHash)) {
      uniqueLogsMap.set(transactionHash, log);
    }
  }

  const currentTimestamp = new Date();

  await Promise.all(
    Array.from(uniqueLogsMap.values()).map(async (log) => {
      const transactionHash = log.transactionHash;

      const block = await client.getBlock({ blockNumber: log.blockNumber, includeTransactions: false });

      if (compareAsc(fromUnixTime(Number(block.timestamp.toString())), subDays(currentTimestamp, 90)) === -1) {
        return;
      }

      const message = await originContract.getMessagesByTransactionHash(transactionHash);
      if (!message || message.length === 0) {
        return;
      }

      const messageStatus = await destinationContract.getMessageStatus(message[0].messageHash);

      const token = tokens.find(
        (token) =>
          token.L1?.toLowerCase() === log.args.token.toLowerCase() ||
          token.L2?.toLowerCase() === log.args.token.toLowerCase(),
      );

      if (!token) {
        return;
      }

      const [amount] = decodeAbiParameters([{ type: "uint256", name: "amount" }], log.data);

      transactionsMap.set(transactionHash, {
        type: BridgeTransactionType.ERC20,
        status: formatStatus(messageStatus),
        token,
        fromChain,
        toChain,
        timestamp: block.timestamp,
        bridgingTx: log.transactionHash,
        message: {
          from: message[0].messageSender as Address,
          to: message[0].destination as Address,
          fee: message[0].fee,
          value: message[0].value,
          nonce: message[0].messageNonce,
          calldata: message[0].calldata,
          messageHash: message[0].messageHash,
          amountSent: amount,
        },
      });
    }),
  );

  return Array.from(transactionsMap.values());
}

async function fetchCCTPBridgeEvents(
  address: Address,
  fromChain: Chain,
  toChain: Chain,
  tokens: Token[],
): Promise<BridgeTransaction[]> {
  const transactionsMap = new Map<string, BridgeTransaction>();
  const fromChainClient = getPublicClient(wagmiConfig, {
    chainId: fromChain.id,
  });
  const toChainClient = getPublicClient(wagmiConfig, {
    chainId: toChain.id,
  });

  const usdcLogs = (await fromChainClient.getLogs({
    event: CCTPDepositForBurnAbiEvent,
    fromBlock: "earliest",
    toBlock: "latest",
    address: fromChain.cctpTokenMessengerV2Address,
    args: {
      depositor: address,
    },
  })) as unknown as DepositForBurnLogEvent[];

  // TODO - Consider deduplication

  const currentTimestamp = new Date();

  // TODO - Minimise # of CCTP API calls in this block
  await Promise.all(
    usdcLogs.map(async (log) => {
      const transactionHash = log.transactionHash;
      // TODO - Search for cache for completed chainId-transactionHash, if cache-hit skip remaining logic

      const fromBlock = await fromChainClient.getBlock({ blockNumber: log.blockNumber, includeTransactions: false });

      if (compareAsc(fromUnixTime(Number(fromBlock.timestamp.toString())), subDays(currentTimestamp, 90)) === -1) {
        return;
      }

      // const token = tokens.find((token) => token.symbol === "USDC" && token.type.includes("bridge-reserved"));
      const token = tokens.find((token) => isCctp(token));
      if (!token) return;

      // TODO - Compute deterministic nonce without consulting CCTP API, to guard against CCTP API rate limit of 10 requests/second
      // TODO - Replace with getCCTPNonce once implemented
      const message = await getCCTPMessageByTxHash(transactionHash, fromChain.cctpDomain);
      if (!message) return;
      const nonce = message.eventNonce;
      // getCCTPNonce(fromChainClient, transactionHash, nonce);

      const isNonceUsed = await isCCTPNonceUsed(toChainClient, nonce, toChain.cctpMessageTransmitterV2Address);

      // TODO - refactor getCCTPTransactionStatus to depend on nonce only, and not on CCTP API response
      const status = getCCTPTransactionStatus(message.status, isNonceUsed);

      // TODO - Save to cache for completed chainId-transactionHash, if cache-hit skip remaining logic
      transactionsMap.set(transactionHash, {
        type: BridgeTransactionType.USDC,
        status,
        token,
        fromChain,
        toChain,
        timestamp: fromBlock.timestamp,
        bridgingTx: log.transactionHash,
        message: {
          amountSent: BigInt(log.args.amount),
          nonce: nonce,
        },
      });
    }),
  );

  return Array.from(transactionsMap.values());
}

function formatStatus(status: OnChainMessageStatus): TransactionStatus {
  switch (status) {
    case OnChainMessageStatus.UNKNOWN:
      return TransactionStatus.PENDING;
    case OnChainMessageStatus.CLAIMABLE:
      return TransactionStatus.READY_TO_CLAIM;
    case OnChainMessageStatus.CLAIMED:
      return TransactionStatus.COMPLETED;
    default:
      return TransactionStatus.PENDING;
  }
}
