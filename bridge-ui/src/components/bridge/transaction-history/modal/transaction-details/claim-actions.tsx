import { useEffect } from "react";

import { useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useConnection, useDisconnect, useSimulateContract, useSwitchChain } from "wagmi";

import { LINEA_ROLLUP_YIELD_EXTENSION_ABI } from "@/abis/LineaRollupYieldExtension";
import Button from "@/components/ui/button";
import { useClaim, useL1MessageServiceLiquidity } from "@/hooks";
import { BridgeTransaction, BridgeTransactionType, ChainLayer } from "@/types";
import { buildClaimWithProofParams, isNativeBridgeMessage } from "@/utils/message";

import styles from "./transaction-details.module.scss";

type ClaimActionsProps = {
  transaction: BridgeTransaction;
  isLoadingClaimTxParams: boolean;
  onCloseModal: () => void;
};

export default function ClaimActions({ transaction, isLoadingClaimTxParams, onCloseModal }: ClaimActionsProps) {
  const { chain, address } = useConnection();
  const { mutate: disconnect } = useDisconnect();
  const { mutate: switchChain, isPending: isSwitchingChain } = useSwitchChain();

  const nativeMessage =
    transaction.message && isNativeBridgeMessage(transaction.message) ? transaction.message : undefined;

  const isL2ToL1EthWithdrawal =
    !!nativeMessage &&
    transaction.fromChain.layer === ChainLayer.L2 &&
    transaction.toChain.layer === ChainLayer.L1 &&
    transaction.type === BridgeTransactionType.ETH;

  const withdrawalAmount = isL2ToL1EthWithdrawal ? nativeMessage.amountSent : 0n;

  const { isLowLiquidity: isMessageServiceBalanceTooLow, isLoading: isLiquidityLoading } = useL1MessageServiceLiquidity(
    {
      toChain: transaction.toChain,
      isL2ToL1Eth: isL2ToL1EthWithdrawal,
      withdrawalAmount,
    },
  );

  const isConnectedWalletMessageRecipient =
    isL2ToL1EthWithdrawal && !!address && nativeMessage.to.toLowerCase() === address.toLowerCase();

  const canAttemptLstClaim =
    isMessageServiceBalanceTooLow && isConnectedWalletMessageRecipient && !!transaction.toChain.yieldProviderAddress;

  const claimWithProofParams =
    canAttemptLstClaim && nativeMessage ? buildClaimWithProofParams(nativeMessage) : undefined;

  const {
    isSuccess: isLstSimulationSuccess,
    isLoading: isLstSimulationLoading,
    error: lstSimulationError,
  } = useSimulateContract({
    address: transaction.toChain.messageServiceAddress,
    abi: LINEA_ROLLUP_YIELD_EXTENSION_ABI,
    functionName: "claimMessageWithProofAndWithdrawLST",
    args: claimWithProofParams
      ? [claimWithProofParams, transaction.toChain.yieldProviderAddress as `0x${string}`]
      : undefined,
    chainId: transaction.toChain.id,
    account: address,
    query: {
      enabled: canAttemptLstClaim && !!claimWithProofParams && !!address,
      retry: 2,
      staleTime: 30_000,
    },
  });

  const canClaimStEth = canAttemptLstClaim && isLstSimulationSuccess;

  const {
    claim,
    isConfirming,
    isPending,
    isConfirmed,
    error: claimError,
  } = useClaim({
    status: transaction.status,
    type: transaction.type,
    fromChain: transaction.fromChain,
    toChain: transaction.toChain,
    args: transaction.message,
    lstSimulationPassed: canClaimStEth,
  });

  const queryClient = useQueryClient();
  useEffect(() => {
    if (isConfirmed) {
      queryClient.invalidateQueries({ queryKey: ["transactionHistory"], exact: false });
      onCloseModal();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isConfirmed]);

  const needsChainSwitch = chain?.id !== transaction.toChain.id;

  const buttonText = (() => {
    if (isLoadingClaimTxParams || isLiquidityLoading) return "Loading Claim Data...";
    if (isPending || isConfirming) return "Waiting for confirmation...";
    if (isSwitchingChain) return "Switching chain...";
    if (isMessageServiceBalanceTooLow && !isConnectedWalletMessageRecipient) return "Switch wallet";
    if (canAttemptLstClaim && !needsChainSwitch && isLstSimulationLoading) return "Simulating claim...";
    if (canAttemptLstClaim && !needsChainSwitch) return "Claim stETH";
    if (needsChainSwitch) return `Switch to ${transaction.toChain.name}`;
    return "Claim";
  })();

  const isButtonDisabled =
    isLoadingClaimTxParams ||
    isLiquidityLoading ||
    isPending ||
    isConfirming ||
    isSwitchingChain ||
    (canAttemptLstClaim && !needsChainSwitch && !isLstSimulationSuccess);

  const handleClaim = () => {
    if (transaction.toChain.id && chain?.id && chain.id !== transaction.toChain.id) {
      switchChain({ chainId: transaction.toChain.id });
      return;
    }
    if (claim) claim();
  };

  const handlePrimaryAction = () => {
    if (isMessageServiceBalanceTooLow && !isConnectedWalletMessageRecipient) {
      disconnect();
      return;
    }
    handleClaim();
  };

  return (
    <>
      <div className={styles.actions}>
        <Button disabled={isButtonDisabled} onClick={handlePrimaryAction} className={styles.action}>
          {buttonText}
        </Button>
        <Button variant="outline" onClick={onCloseModal} className={styles.action}>
          Cancel
        </Button>
      </div>
      {claimError && <p className={styles["error-text"]}>Claim failed. Please try again.</p>}
      {lstSimulationError && !needsChainSwitch && (
        <p className={styles["error-text"]}>
          stETH claiming is currently unavailable. Please wait until stETH or ETH claiming becomes available.
        </p>
      )}
      {isMessageServiceBalanceTooLow && (
        <div className={styles["low-liquidity-info"]}>
          <p className={styles["helper-text"]}>
            {isConnectedWalletMessageRecipient
              ? "Low ETH liquidity. Claim as stETH now or wait until sufficient ETH balance becomes available."
              : "Please connect the recipient wallet to claim stETH, or wait until sufficient ETH balance becomes available."}
          </p>
          {canAttemptLstClaim && (
            <p className={styles["terms-text"]}>
              By claiming, you acknowledge that a liquidity buffer may apply. See{" "}
              <Link href="https://linea.build/terms-of-service" target="_blank" rel="noopener noreferrer">
                Terms & Conditions.
              </Link>
            </p>
          )}
        </div>
      )}
    </>
  );
}
