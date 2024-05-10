'use client';

import { useContext, useEffect, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { useAccount, useWaitForTransaction } from 'wagmi';
import { parseUnits } from 'viem';
import classNames from 'classnames';
import { toast } from 'react-toastify';

import { ChainContext } from '@/contexts/chain.context';
import { useSwitchNetwork, useAllowance, useApprove } from '@/hooks';
import { Transaction } from '@/models';

export type BridgeForm = {
  amount: string;
  balance: string;
  submit: string;
};

export default function Approve() {
  const [waitingTransaction, setWaitingTransaction] = useState<Transaction | undefined>();

  // Form
  const { getValues, setValue, watch } = useFormContext();
  const watchAmount = watch('amount', false);
  const watchBalance = watch('balance', false);

  // Context
  const context = useContext(ChainContext);
  const { token, fromChain, tokenBridgeAddress } = context;

  // Hooks
  const { switchChain } = useSwitchNetwork(fromChain?.id);
  const { allowance, fetchAllowance } = useAllowance();
  const { hash: newTxHash, setHash, writeApprove, isLoading: isApprovalLoading } = useApprove();

  // Wagmi
  const { address } = useAccount();
  const {
    isLoading: isWaitingLoading,
    isSuccess: isWaitingSuccess,
    isError: isWaitingError,
  } = useWaitForTransaction({
    hash: waitingTransaction?.txHash,
    chainId: waitingTransaction?.chainId,
    onSuccess() {
      setWaitingTransaction(undefined);
      setHash(null);
    },
    onError() {
      setWaitingTransaction(undefined);
      setHash(null);
    },
  });

  // Set form allowance
  useEffect(() => {
    setValue('allowance', allowance);
  }, [allowance, setValue]);

  // Set tx hash to trigger useWaitForTransaction
  useEffect(() => {
    newTxHash &&
      setWaitingTransaction({
        txHash: newTxHash,
        chainId: fromChain?.id,
        name: fromChain?.name,
      });
  }, [newTxHash]);

  // Clear tx waiting when changing account
  useEffect(() => {
    setWaitingTransaction(undefined);
  }, [address]);

  // Refresh allowance after successful tx
  useEffect(() => {
    fetchAllowance();
    isWaitingSuccess && toast.success('Token approval successful!');
  }, [isWaitingSuccess, fetchAllowance]);

  useEffect(() => {
    isWaitingError && toast.error('Token approval failed.');
  }, [isWaitingError]);

  useEffect(() => {
    if (token && allowance && watchAmount && parseUnits(watchAmount, token.decimals) <= allowance) {
      setValue('bridgingAllowed', true);
    } else {
      setValue('bridgingAllowed', false);
    }
  }, [watchAmount, allowance, token, setValue]);

  // Click on approve
  const approveHandler = async () => {
    await switchChain();
    if (token) {
      const amount = getValues('amount');
      const amountToApprove = parseUnits(amount, token.decimals);
      writeApprove(amountToApprove, tokenBridgeAddress);
    }
  };

  return (
    <div className="flex flex-col">
      <button
        id="approve-btn"
        className={classNames('btn btn-primary w-48 rounded-full', {
          'btn-disabled':
            token &&
            (isApprovalLoading ||
              isWaitingLoading ||
              !watchAmount ||
              parseUnits(watchAmount, token.decimals) === BigInt(0) ||
              (allowance && parseUnits(watchAmount, token.decimals) <= allowance) ||
              parseUnits(watchAmount, token.decimals) > parseUnits(watchBalance, token.decimals)),
        })}
        type="button"
        onClick={approveHandler}
      >
        {(isApprovalLoading || isWaitingLoading) && <span className="loading loading-spinner"></span>}
        Approve
      </button>
      {/* <div className="mt-5 text-xs">
        Allowed:{" "}
        {allowance && token ? formatUnits(allowance, token.decimals) : ""}{" "}
        {token?.symbol}
      </div> */}
    </div>
  );
}
