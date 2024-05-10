'use client';

import { useContext, useEffect, useRef, useState } from 'react';
import { useForm, FormProvider } from 'react-hook-form';
import { useAccount, useWaitForTransaction } from 'wagmi';
import { parseEther } from 'viem';
import { toast } from 'react-toastify';

import { BridgeForm, Transaction } from '@/models';
import FromChainToChain from './FromChainToChain';
import Amount from './Amount';
import Balance from './Balance';
import { useSwitchNetwork } from '@/hooks';
import Submit from './Submit';
import { ChainContext } from '@/contexts/chain.context';
import Fees from './Fees';
import TokenModal from './TokenModal';
import useBridge from '@/hooks/useBridge';
import Recipient from './Recipient';
import { TokenType } from '@/config/config';
import { HistoryContext } from '@/contexts/history.context';
import { useConfigContext } from '@/contexts/token.context';

export default function Bridge() {
  const { tokensConfig: configContextValue } = useConfigContext();
  const [waitingTransaction, setWaitingTransaction] = useState<Transaction | undefined>();

  // Dialog Ref
  const tokensModalRef = useRef<HTMLDialogElement>(null);

  // Context
  const context = useContext(ChainContext);
  const { fromChain, token } = context;
  const historyContext = useContext(HistoryContext);
  const { fetchHistory } = historyContext;

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
      fetchHistory();
    },
    onError() {
      setWaitingTransaction(undefined);
    },
  });

  // Form
  const methods = useForm<BridgeForm>({
    defaultValues: {
      token: configContextValue?.UNKNOWN[0],
      claim: token?.type === TokenType.ETH ? 'auto' : 'manual',
      amount: '',
    },
  });

  // Hooks
  const { switchChain } = useSwitchNetwork(fromChain?.id);
  const { hash, isLoading, bridge, error: bridgeError } = useBridge();

  // Set tx hash to trigger useWaitForTransaction
  useEffect(() => {
    if (hash) {
      setWaitingTransaction({
        txHash: hash,
        chainId: fromChain?.id,
        name: fromChain?.name,
      });
    }
  }, [hash]);

  // Clear tx waiting when changing account
  useEffect(() => {
    setWaitingTransaction(undefined);
  }, [address]);

  useEffect(() => {
    isWaitingSuccess && waitingTransaction && toast.success(`Transaction validated on ${waitingTransaction?.name}.`);
  }, [isWaitingSuccess, waitingTransaction]);

  useEffect(() => {
    isWaitingError && toast.error('Token bridging failed.');
  }, [isWaitingError]);

  useEffect(() => {
    bridgeError &&
      bridgeError.displayInToast &&
      toast.error(
        <>
          {bridgeError.message}
          <a href={bridgeError.link} target="_blank" className="ml-1 underline">
            here
          </a>
        </>,
        {
          autoClose: false,
          draggable: false,
          closeOnClick: false,
          style: { width: 500, marginLeft: -90 },
        },
      );
  }, [bridgeError]);

  // Click on approve
  const onSubmit = async (data: BridgeForm) => {
    if (isLoading || isWaitingLoading) {
      return;
    }
    await switchChain();
    bridge(data.amount, parseEther(data.minFees), data.recipient);
  };

  return (
    <FormProvider {...methods}>
      <form onSubmit={methods.handleSubmit(onSubmit)}>
        <div className="card-body">
          <h2 className="mb-5 font-medium text-white card-title">Bridge</h2>
          <ul className="space-y-6">
            <li>
              <FromChainToChain />
            </li>
            <li>
              <Amount tokensModalRef={tokensModalRef} />
            </li>
            <li className="text-sm text-end">
              <Balance />
            </li>
            <li>
              <Recipient />
            </li>
            <li>
              <div className="divider"></div>
            </li>
            <li className="text-sm">
              <Fees />
            </li>
            <li>
              <Submit isLoading={isLoading ? true : false} isWaitingLoading={isWaitingLoading} />
            </li>
          </ul>
        </div>
      </form>
      <TokenModal tokensModalRef={tokensModalRef} />
    </FormProvider>
  );
}
