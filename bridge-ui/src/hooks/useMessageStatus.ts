import { useCallback } from "react";
import { OnChainMessageStatus } from "@consensys/linea-sdk";
import useLineaSDK from "./useLineaSDK";
import { MessageWithStatus } from "./useClaimTransaction";
import { useChainStore } from "@/stores/chainStore";
import { ChainLayer } from "@/types";

const useMessageStatus = () => {
  const { lineaSDK, lineaSDKContracts } = useLineaSDK();

  const fromChain = useChainStore.useFromChain();
  const toChain = useChainStore.useToChain();

  const getMessagesByTransactionHash = useCallback(
    async (transactionHash: string) => {
      return await lineaSDKContracts[fromChain.layer]?.getMessagesByTransactionHash(transactionHash);
    },
    [lineaSDKContracts, fromChain.layer],
  );

  const getMessageStatuses = useCallback(
    async (transactionHash: string) => {
      const messages = await getMessagesByTransactionHash(transactionHash);

      const messagesWithStatuses: Array<MessageWithStatus> = [];
      if (messages && messages.length > 0) {
        const otherLayer = toChain.layer;

        const promises = messages.map(async (message) => {
          const l1Chain = fromChain.layer === ChainLayer.L1 ? fromChain : toChain;
          const l2Chain = fromChain.layer === ChainLayer.L2 ? fromChain : toChain;

          const l1ClaimingService = lineaSDK?.getL1ClaimingService(
            l1Chain.messageServiceAddress,
            l2Chain.messageServiceAddress,
          );
          let status: OnChainMessageStatus;
          let claimingTransactionHash;
          // For messages to claim on L1 we check if we need to claim with the new claiming method
          // which requires the proof linked to this message
          let proof;

          if (otherLayer === ChainLayer.L1) {
            // Message from L2 to L1
            status = (await l1ClaimingService?.getMessageStatus(message.messageHash)) || OnChainMessageStatus.UNKNOWN;

            if (
              status === OnChainMessageStatus.CLAIMABLE &&
              (await l1ClaimingService?.isClaimingNeedingProof(message.messageHash))
            ) {
              try {
                proof = await l1ClaimingService?.getMessageProof(message.messageHash);
              } catch (ex) {
                // We ignore the error, the proof will stay undefined, we assume
                // it's a message from the old message service
              }
            }
            if (status === OnChainMessageStatus.CLAIMED) {
              const [messageClaimedEvent] = await lineaSDKContracts.L1.getEvents(
                lineaSDKContracts.L1.contract.filters.MessageClaimed(message.messageHash),
              );
              claimingTransactionHash = messageClaimedEvent ? messageClaimedEvent.transactionHash : undefined;
            }
          } else {
            // Message from L1 to L2
            status = await lineaSDKContracts.L2.getMessageStatus(message.messageHash);
            const [messageClaimedEvent] = await lineaSDKContracts.L2.getEvents(
              lineaSDKContracts.L2.contract.filters.MessageClaimed(message.messageHash),
            );
            claimingTransactionHash = messageClaimedEvent ? messageClaimedEvent.transactionHash : undefined;
          }

          // Convert the BigNumbers to string for serialization issue with the storage
          const messageWithStatus: MessageWithStatus = {
            calldata: message.calldata,
            destination: message.destination,
            fee: message.fee.toString(),
            messageHash: message.messageHash,
            messageNonce: message.messageNonce.toString(),
            messageSender: message.messageSender,
            status: status,
            value: message.value.toString(),
            proof,
            claimingTransactionHash,
          };
          messagesWithStatuses.push(messageWithStatus);
        });
        await Promise.all(promises);
      }

      return messagesWithStatuses;
    },
    [lineaSDKContracts, getMessagesByTransactionHash, lineaSDK, fromChain, toChain],
  );

  return { getMessageStatuses };
};

export default useMessageStatus;
