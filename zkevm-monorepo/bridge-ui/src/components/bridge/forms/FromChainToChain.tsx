'use client';

import { useContext } from 'react';
import { MdOutlineArrowRightAlt, MdOutlineCompareArrows } from 'react-icons/md';

import { ChainContext } from '@/contexts/chain.context';
import { useIsConnected } from '@/hooks';
import ChainLogo from '@/components/widgets/ChainLogo';

export default function FromChainToChain() {
  // Context
  const context = useContext(ChainContext);
  const { fromChain, toChain, switchChain, resetToken } = context;

  // Hooks
  const isConnected = useIsConnected();

  /**
   * Swith chain
   */
  const switchChainHandler = () => {
    resetToken();
    switchChain();
  };

  return (
    <div className="flex flex-row justify-between space-x-2">
      <div className="flex flex-row items-center space-x-2">
        {isConnected && (
          <>
            {fromChain && <ChainLogo chainId={fromChain.id} />} <span>{fromChain && fromChain.name}</span>
            <div>
              <MdOutlineArrowRightAlt />
            </div>
            {toChain && <ChainLogo chainId={toChain.id} />} <span>{toChain && toChain.name}</span>
          </>
        )}
      </div>

      <div>
        <button
          id="switch-chain-btn"
          className="btn btn-circle btn-sm btn-info"
          type="button"
          onClick={switchChainHandler}
          disabled={!isConnected}
        >
          <MdOutlineCompareArrows className="text-lg" />
        </button>
      </div>
    </div>
  );
}
