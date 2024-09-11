import { useState, useCallback } from "react";
import log from "loglevel";
import { simulateContract, writeContract } from "@wagmi/core";
import { zeroAddress } from "viem";
import MessageService from "@/abis/MessageService.json";
import { config, wagmiConfig } from "@/config";
import { OnChainMessageStatus } from "@consensys/linea-sdk";
import { Proof } from "@consensys/linea-sdk/dist/lib/sdk/merkleTree/types";
import { TransactionHistory } from "@/models/history";
import { getChainNetworkLayer, getChainNetworkType } from "@/utils/chainsUtil";
import { useAccount } from "wagmi";
import { Transaction } from "@/models";

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
  claimingTransactionHash?: string;
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

const useTransactionManagement = () => {
  const { address } = useAccount();
  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

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

  return { transaction, isLoading, isError: error !== null, error, writeClaimMessage };
};

export default useTransactionManagement;
