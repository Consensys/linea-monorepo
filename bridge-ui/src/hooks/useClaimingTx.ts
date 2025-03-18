import { BridgeTransactionType, CCTPV2BridgeMessage, Chain, NativeBridgeMessage, TransactionStatus } from "@/types";
import { getPublicClient } from "@wagmi/core";
import { config as wagmiConfig } from "@/lib/wagmi";
import { eventCCTPMessageReceived, eventMessageClaimed } from "@/utils/events";
import { isNativeBridgeMessage, isCCTPV2BridgeMessage } from "@/utils/message";
import { useQuery } from "@tanstack/react-query";

type UseClaimingTxProps = {
  status?: TransactionStatus;
  type?: BridgeTransactionType;
  toChain?: Chain;
  args?: NativeBridgeMessage | CCTPV2BridgeMessage;
  bridgingTx?: string;
};

const useClaimingTx = ({ status, type, toChain, args, bridgingTx }: UseClaimingTxProps): string | undefined => {
  // TODO - consider refactor into own file
  // queryFn for useQuery cannot return undefined - https://tanstack.com/query/latest/docs/framework/react/reference/useQuery
  async function getClaimTx(params: {
    status?: TransactionStatus;
    type?: BridgeTransactionType;
    toChain?: Chain;
    args?: NativeBridgeMessage | CCTPV2BridgeMessage;
  }): Promise<string> {
    if (!status || !type || !toChain || !args) return "";
    if (status === TransactionStatus.PENDING) return "";
    const toChainClient = getPublicClient(wagmiConfig, {
      chainId: toChain.id,
    });

    switch (type) {
      case BridgeTransactionType.ETH: {
        if (!isNativeBridgeMessage(args)) return "";
        const messageClaimedEvents = await toChainClient.getLogs({
          event: eventMessageClaimed,
          // TODO - Find more efficient `fromBlock` param than 'earliest'
          fromBlock: "earliest",
          toBlock: "latest",
          address: toChain.messageServiceAddress,
          args: {
            _messageHash: args?.messageHash as `0x${string}`,
          },
        });
        if (messageClaimedEvents.length === 0) return "";
        return messageClaimedEvents[0].transactionHash;
      }
      case BridgeTransactionType.ERC20: {
        if (!isNativeBridgeMessage(args)) return "";
        const messageClaimedEvents = await toChainClient.getLogs({
          event: eventMessageClaimed,
          // TODO - Find more efficient `fromBlock` param than 'earliest'
          fromBlock: "earliest",
          toBlock: "latest",
          address: toChain.messageServiceAddress,
          args: {
            _messageHash: args?.messageHash as `0x${string}`,
          },
        });
        if (messageClaimedEvents.length === 0) return "";
        return messageClaimedEvents[0].transactionHash;
      }
      case BridgeTransactionType.USDC: {
        if (!isCCTPV2BridgeMessage(args) || !args.nonce) return "";
        const messageReceivedEvents = await toChainClient.getLogs({
          event: eventCCTPMessageReceived,
          // TODO - Find more efficient `fromBlock` param than 'earliest'
          fromBlock: "earliest",
          toBlock: "latest",
          address: toChain.cctpMessageTransmitterV2Address,
          args: {
            nonce: args?.nonce,
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
    queryKey: ["useClaimingTx", bridgingTx, toChain?.id],
    queryFn: async () => getClaimTx({ status, type, toChain, args }),
  });

  if (!data || data === "") return;
  return data;
};

export default useClaimingTx;
