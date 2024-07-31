"use client";

import { MdOutlineArrowRightAlt, MdOutlineCompareArrows } from "react-icons/md";
import { useAccount } from "wagmi";
import ChainLogo from "@/components/widgets/ChainLogo";
import { useChainStore } from "@/stores/chainStore";

export default function FromChainToChain() {
  // Context
  const { fromChain, toChain, switchChain, resetToken } = useChainStore((state) => ({
    fromChain: state.fromChain,
    toChain: state.toChain,
    switchChain: state.switchChain,
    resetToken: state.resetToken,
  }));

  // Hooks
  const { isConnected } = useAccount();

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
          className="btn btn-circle btn-info btn-sm"
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
