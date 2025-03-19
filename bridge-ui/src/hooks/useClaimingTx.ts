import { BridgeTransaction, BridgeTransactionType, TransactionStatus, CCTPMessageReceivedAbiEvent } from "@/types";
import { getPublicClient } from "@wagmi/core";
import { config as wagmiConfig } from "@/lib/wagmi";
import { isNativeBridgeMessage, isCCTPV2BridgeMessage } from "@/utils/message";
import { useQuery } from "@tanstack/react-query";
import { MessageClaimedABIEvent } from "@/types";

const useClaimingTx = (transaction: BridgeTransaction | undefined): string | undefined => {
  // TODO - consider refactor into own file
  // queryFn for useQuery cannot return undefined - https://tanstack.com/query/latest/docs/framework/react/reference/useQuery
  async function getClaimTx(transaction: BridgeTransaction | undefined): Promise<string> {
    if (!transaction) return "";
    if (transaction?.claimingTx) return "";
    const { status, type, toChain, message } = transaction;
    if (!status || !type || !toChain || !message) return "";
    if (status === TransactionStatus.PENDING) return "";

    const toChainClient = getPublicClient(wagmiConfig, {
      chainId: toChain.id,
    });

    switch (type) {
      case BridgeTransactionType.ETH: {
        if (!isNativeBridgeMessage(message)) return "";
        const messageClaimedEvents = await toChainClient.getLogs({
          event: MessageClaimedABIEvent,
          // TODO - Find more efficient `fromBlock` param than 'earliest'
          fromBlock: "earliest",
          toBlock: "latest",
          address: toChain.messageServiceAddress,
          args: {
            _messageHash: message?.messageHash as `0x${string}`,
          },
        });
        if (messageClaimedEvents.length === 0) return "";
        const a = messageClaimedEvents[0].args;
        return messageClaimedEvents[0].transactionHash;
      }
      case BridgeTransactionType.ERC20: {
        if (!isNativeBridgeMessage(message)) return "";
        const messageClaimedEvents = await toChainClient.getLogs({
          event: MessageClaimedABIEvent,
          // TODO - Find more efficient `fromBlock` param than 'earliest'
          fromBlock: "earliest",
          toBlock: "latest",
          address: toChain.messageServiceAddress,
          args: {
            _messageHash: message?.messageHash as `0x${string}`,
          },
        });
        if (messageClaimedEvents.length === 0) return "";
        return messageClaimedEvents[0].transactionHash;
      }
      case BridgeTransactionType.USDC: {
        if (!isCCTPV2BridgeMessage(message) || !message.nonce) return "";
        const messageReceivedEvents = await toChainClient.getLogs({
          event: CCTPMessageReceivedAbiEvent,
          // TODO - Find more efficient `fromBlock` param than 'earliest'
          fromBlock: "earliest",
          toBlock: "latest",
          address: toChain.cctpMessageTransmitterV2Address,
          args: {
            nonce: message?.nonce,
          },
        });
        if (messageReceivedEvents.length === 0) return "";
        return messageReceivedEvents[0].transactionHash;
      }
      default: {
        return "";
      }
    }
  }

  const { data } = useQuery({
    // TODO - Do we need to account for undefined props here? Otherwise caching behaviour is not as expected?
    queryKey: ["useClaimingTx", transaction?.bridgingTx, transaction?.toChain?.id],
    queryFn: async () => getClaimTx(transaction),
  });

  if (!data || data === "") return;
  return data;
};

export default useClaimingTx;
