import { Address } from "viem";
import { compareAsc, fromUnixTime, subDays } from "date-fns";
import { getPublicClient } from "@wagmi/core";
import { config as wagmiConfig } from "@/lib/wagmi";
import { BridgeTransaction, BridgeTransactionType, Chain, Token, CCTPDepositForBurnAbiEvent } from "@/types";
import { isCctp, isCCTPNonceUsed, getCCTPTransactionStatus, getCCTPMessageByTxHash } from "@/utils";
import { DepositForBurnLogEvent } from "@/types/events";

export async function fetchCCTPBridgeEvents(
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
