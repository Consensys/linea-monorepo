import { useCallback } from "react";
import { OnChainMessageStatus } from "@consensys/linea-sdk";
import useLineaSDK from "./useLineaSDK";
import { config, NetworkLayer } from "@/config";
import { MessageWithStatus } from "./useClaimTransaction";
import { useChainStore } from "@/stores/chainStore";

const useMessageStatus = () => {
  const { lineaSDK, lineaSDKContracts } = useLineaSDK();

  const networkType = useChainStore((state) => state.networkType);

  const getMessagesByTransactionHash = useCallback(
    async (transactionHash: string, networkLayer: NetworkLayer) => {
      if (!lineaSDKContracts || networkLayer === NetworkLayer.UNKNOWN) {
        return;
      }
      return await lineaSDKContracts[networkLayer]?.getMessagesByTransactionHash(transactionHash);
    },
    [lineaSDKContracts],
  );

  const getMessageStatuses = useCallback(
    async (transactionHash: string, networkLayer: NetworkLayer) => {
      if (!lineaSDKContracts || networkLayer === NetworkLayer.UNKNOWN) {
        return;
      }

      const messages = await getMessagesByTransactionHash(transactionHash, networkLayer);

      const messagesWithStatuses: Array<MessageWithStatus> = [];
      if (messages && messages.length > 0) {
        const otherLayer = networkLayer === NetworkLayer.L1 ? NetworkLayer.L2 : NetworkLayer.L1;

        const promises = messages.map(async (message) => {
          const l1ClaimingService = lineaSDK?.getL1ClaimingService(
            config.networks[networkType].L1.messageServiceAddress,
            config.networks[networkType].L2.messageServiceAddress,
          );
          let status: OnChainMessageStatus;
          let claimingTransactionHash;
          // For messages to claim on L1 we check if we need to claim with the new claiming method
          // which requires the proof linked to this message
          let proof;

          if (otherLayer === NetworkLayer.L1) {
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
    [lineaSDKContracts, getMessagesByTransactionHash, lineaSDK, networkType],
  );

  return { getMessageStatuses };
};

export default useMessageStatus;
