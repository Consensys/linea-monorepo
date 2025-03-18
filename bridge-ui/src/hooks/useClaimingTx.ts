import { BridgeTransactionType, CCTPV2BridgeMessage, Chain, NativeBridgeMessage } from "@/types";
import { getPublicClient } from "@wagmi/core";
import { config as wagmiConfig } from "@/lib/wagmi";
import { eventCCTPMessageReceived } from "@/utils/events";
import { isNativeBridgeMessage, isCCTPV2BridgeMessage } from "@/utils/message";
import { useQuery } from "@tanstack/react-query";

type UseClaimingTxProps = {
  type?: BridgeTransactionType;
  toChain?: Chain;
  args?: NativeBridgeMessage | CCTPV2BridgeMessage;
  bridgingTx?: string;
};

const useClaimingTx = ({ type, toChain, args, bridgingTx }: UseClaimingTxProps) => {
  if (!type || !toChain || !args || !bridgingTx) return;

  // const toChainClient = getPublicClient(wagmiConfig, {
  //   chainId: toChain.id,
  // });

  // let claimingTx: string | undefined;

  // switch (type) {
  //   case BridgeTransactionType.ETH:
  //     if (!isNativeBridgeMessage(args)) return;
  //     break;
  //   case BridgeTransactionType.ERC20:
  //     if (!isNativeBridgeMessage(args)) return;
  //     break;
  //   case BridgeTransactionType.USDC:
  //     if (!isCCTPV2BridgeMessage(args) || !args.nonce) return;
  //     const getMessageReceivedEvents = async (): Promise<string | undefined> => {
  //       const messageReceivedEvents = await toChainClient.getLogs({
  //         event: eventCCTPMessageReceived,
  //         // TODO - Find more efficient `fromBlock` param than 'earliest'
  //         fromBlock: "earliest",
  //         toBlock: "latest",
  //         address: toChain.cctpMessageTransmitterV2Address,
  //         args: {
  //           nonce: args?.nonce,
  //         },
  //       });
  //       if (messageReceivedEvents.length === 0) return undefined;
  //       return messageReceivedEvents[0].transactionHash;
  //     };

  //     const { data } = useQuery({
  //       queryKey: ["useClaimingTx", bridgingTx, toChain.id],
  //     });

  //     break;
  //   default:
  //     return;
  // }

  // return claimingTx;
};

export default useClaimingTx;
