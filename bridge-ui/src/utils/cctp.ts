import MessageTransmitterV2 from "@/abis/MessageTransmitterV2.json";
import { TransactionStatus } from "@/types";
import { CctpAttestationApiResponse, CctpAttestationMessageStatus } from "@/types/cctp";
import { GetPublicClientReturnType } from "@wagmi/core";
import { getAddress } from "viem";

// TODO - Make dynamic based on user actions and current chain, rather than hard-coded
export const CCTP_TOKEN_MESSENGER = getAddress("0x8FE6B999Dc680CcFDD5Bf7EB0974218be2542DAA");
export const CCTP_TRANSFER_MAX_FEE = 500n;
export const CCTP_MIN_FINALITY_THRESHOLD = 1000; // 1000 Fast transfer, 2000 Standard transfer
export const CCTP_MESSAGE_TRANSMITTER = getAddress("0xE737e5cEBEEBa77EFE34D4aa090756590b1CE275");

// See CCTPV2 message encoding scheme at https://github.com/circlefin/evm-cctp-contracts/blob/6e7513cdb2bee6bb0cddf331fe972600fc5017c9/src/messages/v2/MessageV2.sol#L31-L41
export const getCCTPMessageNonce = (message: string): string => {
  if (message.length < 90) return "0x";
  return "0x" + message.substring(26, 90);
};

export const isCCTPNonceUsed = async (client: GetPublicClientReturnType, nonce: string): Promise<boolean> => {
  const resp = await client?.readContract({
    address: CCTP_MESSAGE_TRANSMITTER,
    abi: MessageTransmitterV2.abi,
    functionName: "usedNonces",
    args: [nonce],
  });

  return resp === 1n;
};

export const getCCTPTransactionStatus = (
  cctpMessageStatus: CctpAttestationMessageStatus,
  isNonceUsed: boolean,
): TransactionStatus => {
  if (cctpMessageStatus === "pending_confirmations") return TransactionStatus.PENDING;
  if (!isNonceUsed) return TransactionStatus.READY_TO_CLAIM;
  return TransactionStatus.COMPLETED;
};
