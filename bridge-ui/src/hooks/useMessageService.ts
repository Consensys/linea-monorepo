import { useState, useCallback, useMemo } from "react";
import { readContract, simulateContract, writeContract } from "@wagmi/core";
import { useAccount } from "wagmi";
import { zeroAddress } from "viem";
import log from "loglevel";
import { LineaSDK, OnChainMessageStatus } from "@consensys/linea-sdk";
import { L1MessageServiceContract, L2MessageServiceContract } from "@consensys/linea-sdk/dist/lib/contracts";
import MessageService from "@/abis/MessageService.json";
import { getChainNetworkLayer, getChainNetworkType } from "@/utils/chainsUtil";
import { NetworkLayer, NetworkType, config, wagmiConfig } from "@/config";
import { Transaction } from "@/models";
import { TransactionHistory } from "@/models/history";
import { Proof } from "@consensys/linea-sdk/dist/lib/sdk/merkleTree/types";
import { useChainStore } from "@/stores/chainStore";

interface LineaSDKContracts {
  L1: L1MessageServiceContract;
  L2: L2MessageServiceContract;
}

export interface MessageWithStatus {
  status: OnChainMessageStatus;
  messageSender: string;
  destination: string;
  fee: string;
  value: string;
  messageNonce: string;
  calldata: string;
  messageHash: string;
  proof: Proof | undefined;
}

interface ClaimMessageWithProofParams {
  proof: string[];
  messageNumber: string;
  leafIndex: number;
  from: string;
  to: string;
  fee: string;
  value: string;
  feeRecipient: string;
  merkleRoot: string;
  data: string;
}

const useMessageService = () => {
  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [minimumFee, setMinimumFee] = useState(BigInt(0));
  const [lineaSDK, setLineaSDK] = useState<LineaSDK | undefined>(undefined);
  const [lineaSDKContracts, setLineaSDKContracts] = useState<LineaSDKContracts | undefined>(undefined);

  const { messageServiceAddress, fromChain, networkType } = useChainStore((state) => ({
    messageServiceAddress: state.messageServiceAddress,
    fromChain: state.fromChain,
    networkType: state.networkType,
  }));
  const { address } = useAccount();

  useMemo(() => {
    let _lineaSDK;
    const infuraKey = process.env.NEXT_PUBLIC_INFURA_ID;
    switch (networkType) {
      case NetworkType.MAINNET:
        _lineaSDK = new LineaSDK({
          l1RpcUrl: `https://mainnet.infura.io/v3/${infuraKey}`,
          l2RpcUrl: `https://linea-mainnet.infura.io/v3/${infuraKey}`,
          network: "linea-mainnet",
          mode: "read-only",
        });

        break;
      case NetworkType.SEPOLIA:
        _lineaSDK = new LineaSDK({
          l1RpcUrl: `https://sepolia.infura.io/v3/${infuraKey}`,
          l2RpcUrl: `https://linea-sepolia.infura.io/v3/${infuraKey}`,
          network: "linea-sepolia",
          mode: "read-only",
        });
        break;
    }
    if (!_lineaSDK) {
      return;
    }
    const newLineaSDKContracts: LineaSDKContracts = {
      L1: _lineaSDK.getL1Contract(),
      L2: _lineaSDK.getL2Contract(),
    };
    setLineaSDKContracts(newLineaSDKContracts);
    setLineaSDK(_lineaSDK);
  }, [networkType]);

  useMemo(() => {
    const readMinimumFee = async () => {
      setError(null);
      setIsLoading(true);

      if (!messageServiceAddress) {
        return;
      }

      try {
        let fees = BigInt(0);
        if (fromChain && getChainNetworkLayer(fromChain) === NetworkLayer.L2) {
          //Get the minimum to send along the message for L2
          fees = (await readContract(wagmiConfig, {
            address: messageServiceAddress,
            abi: MessageService.abi,
            functionName: "minimumFeeInWei",
            chainId: fromChain.id,
          })) as bigint;
        }
        setMinimumFee(fees);
      } catch (error) {
        setError(error as Error);
      }

      setIsLoading(false);
    };

    readMinimumFee();
  }, [messageServiceAddress, fromChain]);

  const getMessagesByTransactionHash = useCallback(
    async (transactionHash: string, networkLayer: NetworkLayer) => {
      if (!lineaSDKContracts || networkLayer === NetworkLayer.UNKNOWN) {
        return;
      }
      return await lineaSDKContracts[networkLayer]?.getMessagesByTransactionHash(transactionHash);
    },
    [lineaSDKContracts],
  );

  const getMessagesStatusesByTransactionHash = useCallback(
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
          let status;
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
          } else {
            // Message from L1 to L2
            status = await lineaSDKContracts.L2.getMessageStatus(message.messageHash);
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
          };
          messagesWithStatuses.push(messageWithStatus);
        });
        await Promise.all(promises);
      }

      return messagesWithStatuses;
    },
    [lineaSDKContracts, getMessagesByTransactionHash, lineaSDK, networkType],
  );

  const writeClaimMessage = useCallback(
    async (message: MessageWithStatus, tx: TransactionHistory) => {
      setError(null);
      setIsLoading(true);

      // Get the right message Service address depending on the transaction
      const txNetworkLayer = getChainNetworkLayer(tx.toChain);
      const txNetworkType = getChainNetworkType(tx.toChain);

      if (address && txNetworkLayer && txNetworkType) {
        try {
          const { messageSender, destination, calldata, fee, messageNonce, value, proof } = message;

          const messageServiceAddress = config.networks[txNetworkType][txNetworkLayer].messageServiceAddress;
          if (messageServiceAddress === null) {
            return;
          }
          let writeConfig;
          if (!proof) {
            // Claiming using old message service
            writeConfig = await simulateContract(wagmiConfig, {
              address: messageServiceAddress,
              abi: MessageService.abi,
              functionName: "claimMessage",
              args: [messageSender, destination, fee, value, zeroAddress, calldata, messageNonce],
              chainId: tx.toChain.id,
            });
          } else {
            // Claiming on L1 with new message service
            const params: ClaimMessageWithProofParams = {
              data: calldata,
              fee,
              feeRecipient: zeroAddress,
              from: messageSender,
              to: destination,
              leafIndex: proof.leafIndex,
              merkleRoot: proof.root,
              messageNumber: messageNonce,
              proof: proof.proof,
              value,
            };
            writeConfig = await simulateContract(wagmiConfig, {
              address: messageServiceAddress,
              abi: MessageService.abi,
              functionName: "claimMessageWithProof",
              args: [params],
              chainId: tx.toChain.id,
            });
          }

          const hash = await writeContract(wagmiConfig, writeConfig.request);

          setTransaction({
            txHash: hash,
            chainId: tx.fromChain.id,
            name: tx.fromChain.name,
          });
        } catch (error) {
          log.error(error);
          setError(error as Error);
          setTransaction(null);
        }
      }

      setIsLoading(false);
    },
    [address],
  );

  return {
    isLoading,
    isError: error !== null,
    error,
    minimumFee,
    transaction,
    getMessagesStatusesByTransactionHash,
    writeClaimMessage,
  };
};

export default useMessageService;
