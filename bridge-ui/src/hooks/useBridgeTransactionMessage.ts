import { BridgeTransaction, BridgeTransactionType, CCTPV2BridgeMessage, NativeBridgeMessage } from "@/types";
import { getPublicClient } from "@wagmi/core";
import { config as wagmiConfig } from "@/lib/wagmi";
import { isCCTPV2BridgeMessage } from "@/utils/message";
import { useQuery } from "@tanstack/react-query";
import { getCCTPMessageByNonce, refreshCCTPMessageIfNeeded } from "@/utils";

const useBridgeTransactionMessage = (
  transaction: BridgeTransaction | undefined,
): CCTPV2BridgeMessage | NativeBridgeMessage | undefined => {
  // TODO - consider refactor into own file
  // queryFn for useQuery cannot return undefined - https://tanstack.com/query/latest/docs/framework/react/reference/useQuery
  async function getBridgeTransactionMessage(
    transaction: BridgeTransaction | undefined,
  ): Promise<CCTPV2BridgeMessage | NativeBridgeMessage> {
    const { status, type, fromChain, toChain, message } = transaction as BridgeTransaction;
    if (!status || !type || !fromChain || !toChain || !message) return message;

    // const fromChainClient = getPublicClient(wagmiConfig, {
    //   chainId: fromChain.id,
    // });

    const toChainClient = getPublicClient(wagmiConfig, {
      chainId: toChain.id,
    });

    switch (type) {
      case BridgeTransactionType.ETH: {
        return message;
      }
      case BridgeTransactionType.ERC20: {
        return message;
      }
      case BridgeTransactionType.USDC: {
        if (!isCCTPV2BridgeMessage(message) || !message?.nonce) return message;
        const { nonce } = message;
        // Get message + attestation from CCTP API
        const cctpApiResp = await getCCTPMessageByNonce(nonce, fromChain.cctpDomain);
        if (!cctpApiResp) return message;

        // If expired, get new message + attestation
        const refreshedMessage = await refreshCCTPMessageIfNeeded(
          cctpApiResp,
          status,
          await toChainClient.getBlockNumber(),
          fromChain.cctpDomain,
        );
        if (!refreshedMessage) return message;

        // Populate message + attestation fields
        return {
          message: refreshedMessage.message,
          attestation: refreshedMessage.attestation,
          amountSent: message.amountSent,
          nonce: message.nonce,
        };
      }
      default: {
        return message;
      }
    }
  }

  const { data } = useQuery({
    // TODO - Do we need to account for undefined props here? Otherwise caching behaviour is not as expected?
    queryKey: ["useBridgeTransactionMessage", transaction?.bridgingTx, transaction?.toChain?.id],
    queryFn: async () => getBridgeTransactionMessage(transaction),
  });

  return data;
};

export default useBridgeTransactionMessage;
