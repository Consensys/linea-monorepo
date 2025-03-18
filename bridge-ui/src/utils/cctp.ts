import MessageTransmitterV2 from "@/abis/MessageTransmitterV2.json";
import {
  CCTPMessageReceivedEvent,
  CctpAttestationMessage,
  CctpAttestationMessageStatus,
  TransactionStatus,
} from "@/types";
import { GetPublicClientReturnType } from "@wagmi/core";
import { eventCCTPMessageReceived } from "./events";
import {
  fetchCctpAttestationByNonce,
  fetchCctpAttestationByTxHash,
  reattestCCTPV2PreFinalityMessage,
} from "@/services/cctp";

// TODO - Find optimal value
export const CCTP_TRANSFER_MAX_FEE = 500n;
// 1000 Fast transfer, 2000 Standard transfer
export const CCTP_MIN_FINALITY_THRESHOLD = 1000;

export const isCCTPNonceUsed = async (
  client: GetPublicClientReturnType,
  nonce: string,
  cctpMessageTransmitterV2Address: `0x${string}`,
): Promise<boolean> => {
  const resp = await client?.readContract({
    address: cctpMessageTransmitterV2Address,
    abi: MessageTransmitterV2.abi,
    functionName: "usedNonces",
    args: [nonce],
  });

  return resp === 1n;
};

export const getCCTPMessageExpiryBlock = (message: string): bigint => {
  // See CCTPV2 message format at https://developers.circle.com/stablecoins/message-format
  const expiryInHex = message.substring(2 + 344 * 2, 2 + 376 * 2);
  // Return bigint because this is also returned by Viem client.getBlockNumber()
  return BigInt(parseInt(expiryInHex, 16));
};

export const getCCTPTransactionStatus = (
  cctpMessageStatus: CctpAttestationMessageStatus,
  isNonceUsed: boolean,
): TransactionStatus => {
  if (cctpMessageStatus === "pending_confirmations") return TransactionStatus.PENDING;
  if (!isNonceUsed) return TransactionStatus.READY_TO_CLAIM;
  return TransactionStatus.COMPLETED;
};

export const refreshCCTPMessageIfNeeded = async (
  message: CctpAttestationMessage,
  isNonceUsed: boolean,
  currentToBlock: bigint,
  fromChainCCTPDomain: number,
): Promise<CctpAttestationMessage | undefined> => {
  if (isNonceUsed) return message;
  // Assume that 'pending_confirmations' implies that a message will not be expired
  if (message.status === "pending_confirmations") return message;

  // Check expiry of current message
  const expiryBlock = getCCTPMessageExpiryBlock(message.message);
  if (expiryBlock === 0n) return message;
  if (currentToBlock < expiryBlock) return message;

  // We have an expired message, reattest
  // TODO - Investigate if this will result in an edge case where a 'READY_TO_CLAIM' tx regresses to a 'PENDING' tx
  await reattestCCTPV2PreFinalityMessage(message.eventNonce);

  const refreshedMessage = await getCCTPMessageByNonce(message.eventNonce, fromChainCCTPDomain);
  return refreshedMessage;
};

export const getCCTPMessageByTxHash = async (
  transactionHash: string,
  fromChainCCTPDomain: number,
): Promise<CctpAttestationMessage | undefined> => {
  const attestationApiResp = await fetchCctpAttestationByTxHash(fromChainCCTPDomain, transactionHash);
  if (!attestationApiResp) return;
  const message = attestationApiResp.messages[0];
  if (!message) return;
  return message;
};

const getCCTPMessageByNonce = async (
  nonce: string,
  fromChainCCTPDomain: number,
): Promise<CctpAttestationMessage | undefined> => {
  const attestationApiResp = await fetchCctpAttestationByNonce(fromChainCCTPDomain, nonce);
  if (!attestationApiResp) return;
  const message = attestationApiResp.messages[0];
  if (!message) return;
  return message;
};
