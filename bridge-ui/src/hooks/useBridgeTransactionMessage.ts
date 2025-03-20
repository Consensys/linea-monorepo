import {
  BridgeTransaction,
  BridgeTransactionType,
  CCTPV2BridgeMessage,
  ChainLayer,
  NativeBridgeMessage,
  TransactionStatus,
} from "@/types";
import { getPublicClient } from "@wagmi/core";
import { config as wagmiConfig } from "@/lib/wagmi";
import { isCCTPV2BridgeMessage, isNativeBridgeMessage } from "@/utils/message";
import { useQuery } from "@tanstack/react-query";
import { getCCTPMessageByNonce, refreshCCTPMessageIfNeeded } from "@/utils";
import useLineaSDK from "./useLineaSDK";

const useBridgeTransactionMessage = (
  transaction: BridgeTransaction | undefined,
): { message: CCTPV2BridgeMessage | NativeBridgeMessage | undefined; isLoading: boolean } => {
  const { lineaSDK } = useLineaSDK();

  // TODO - consider refactor into own file
  // queryFn for useQuery cannot return undefined - https://tanstack.com/query/latest/docs/framework/react/reference/useQuery
  async function getBridgeTransactionMessage(
    transaction: BridgeTransaction | undefined,
  ): Promise<CCTPV2BridgeMessage | NativeBridgeMessage> {
    const { status, type, fromChain, toChain, message } = transaction as BridgeTransaction;
    if (!status || !type || !fromChain || !toChain || !message) return message;
    // Cannot claim, so don't waste time getting claim parameters
    if (status !== TransactionStatus.READY_TO_CLAIM) return message;

    // TODO - Refactor each case's logic into its own file
    switch (type) {
      case BridgeTransactionType.ETH: {
        if (toChain.layer === ChainLayer.L2) return message;
        if (!isNativeBridgeMessage(message) || !message?.messageHash) return message;
        const { messageHash } = message;
        message.proof = await lineaSDK.getL1ClaimingService().getMessageProof(messageHash);
        return message;
      }
      case BridgeTransactionType.ERC20: {
        if (toChain.layer === ChainLayer.L2) return message;
        if (!isNativeBridgeMessage(message) || !message?.messageHash) return message;
        const { messageHash } = message;
        message.proof = await lineaSDK.getL1ClaimingService().getMessageProof(messageHash);
        return message;
      }
      case BridgeTransactionType.USDC: {
        if (!isCCTPV2BridgeMessage(message) || !message?.nonce) return message;
        const { nonce } = message;
        // Get message + attestation from CCTP API if we have not already
        // We should have queried CCTP API previously in fetchCCTPBridgeEvents, so should not execute this if block
        if (!message.attestation || !message.message) {
          const cctpApiResp = await getCCTPMessageByNonce(nonce, fromChain.cctpDomain, fromChain.testnet);
          if (!cctpApiResp) return message;
          message.message = cctpApiResp.message;
          message.attestation = cctpApiResp.attestation;
        }

        const toChainClient = getPublicClient(wagmiConfig, {
          chainId: toChain.id,
        });

        // If expired, get new message + attestation
        const refreshedMessage = await refreshCCTPMessageIfNeeded(
          message.message as `0x${string}`,
          message.attestation as `0x${string}`,
          status,
          await toChainClient.getBlockNumber(),
          fromChain.cctpDomain,
          nonce,
          fromChain.testnet,
        );
        if (!refreshedMessage) return message;

        // Populate message + attestation fields
        return {
          message: refreshedMessage.message,
          attestation: refreshedMessage.attestation,
          amountSent: message.amountSent,
          nonce: message.nonce,
          isStatusRegression: refreshedMessage.isStatusRegression,
        };
      }
      default: {
        return message;
      }
    }
  }

  const { data, isLoading } = useQuery({
    // TODO - Do we need to account for undefined props here? Otherwise caching behaviour is not as expected?
    queryKey: ["useBridgeTransactionMessage", transaction?.bridgingTx, transaction?.toChain?.id],
    queryFn: async () => getBridgeTransactionMessage(transaction),
  });

  return { message: data, isLoading };
};

export default useBridgeTransactionMessage;
