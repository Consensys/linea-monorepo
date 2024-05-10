'use client';

import { useContext, useEffect, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { useAccount, useBalance } from 'wagmi';
import { FetchBalanceResult } from '@wagmi/core';
import classNames from 'classnames';

import { formatBalance } from '@/utils/format';
import { ChainContext } from '@/contexts/chain.context';
import { useIsConnected } from '@/hooks';

export default function Balance() {
  const [currentBalance, setCurrentBalance] = useState<FetchBalanceResult | undefined>();
  // Context
  const context = useContext(ChainContext);
  const { token, networkLayer, fromChain } = context;

  const tokenAddress = token && token[networkLayer] ? token[networkLayer] : undefined;

  // Wagmi
  const { address } = useAccount();
  const { data: balance } = useBalance({
    address,
    token: tokenAddress ?? undefined,
    chainId: fromChain?.id,
    watch: true,
  });

  // Hooks
  const isConnected = useIsConnected();

  // Form
  const { setValue } = useFormContext();

  useEffect(() => {
    if (balance) {
      setValue('balance', balance.formatted);
      setCurrentBalance(balance);
    } else {
      setValue('balance', '');
      setCurrentBalance(undefined);
    }
  }, [balance, setValue]);

  return (
    <div
      className={classNames('', {
        'text-neutral-600': !isConnected,
      })}
    >
      Balance: {formatBalance(currentBalance?.formatted)} {currentBalance?.symbol}
    </div>
  );
}
