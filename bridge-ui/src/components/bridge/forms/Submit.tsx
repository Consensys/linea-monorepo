'use client';

import { useContext } from 'react';
import { useFormContext } from 'react-hook-form';
import classNames from 'classnames';

import { ChainContext, NetworkLayer } from '@/contexts/chain.context';
import ApproveERC20 from './ApproveERC20';
import useBridge from '@/hooks/useBridge';

interface Props {
  isLoading: boolean;
  isWaitingLoading: boolean;
}

export default function Submit({ isLoading = false, isWaitingLoading = false }: Props) {
  // Form
  const { watch, formState } = useFormContext();
  const { errors } = formState;

  const watchAmount = watch('amount', false);
  const watchAllowance = watch('allowance', false);

  // Wagmi
  const { bridgeEnabled } = useBridge();

  // Context
  const context = useContext(ChainContext);
  const { token, networkLayer } = context;

  return (
    <div>
      {token && networkLayer !== NetworkLayer.UNKNOWN && token[networkLayer] ? (
        <div className="flex flex-row justify-between">
          <ApproveERC20 />
          <button
            id="submit-erc-btn"
            className={classNames('btn btn-primary w-48 rounded-full', {
              'cursor-wait': isLoading || isWaitingLoading,
              'btn-disabled': !bridgeEnabled(watchAmount, watchAllowance, errors),
            })}
            type="submit"
          >
            {(isLoading || isWaitingLoading) && <span className="loading loading-spinner"></span>}
            Start bridging
          </button>
        </div>
      ) : (
        <button
          id="submit-eth-btn"
          className={classNames('btn w-full btn-primary rounded-full', {
            'cursor-wait': isLoading || isWaitingLoading,
            'btn-disabled': !bridgeEnabled(watchAmount, BigInt(0), errors),
          })}
          type="submit"
        >
          {(isLoading || isWaitingLoading) && <span className="loading loading-spinner"></span>}
          Start bridging
        </button>
      )}
    </div>
  );
}
