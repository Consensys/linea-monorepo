"use client";

import { useFormContext } from "react-hook-form";
import classNames from "classnames";
import ApproveERC20 from "./ApproveERC20";
import useBridge from "@/hooks/useBridge";
import { NetworkLayer } from "@/config";
import { useChainStore } from "@/stores/chainStore";

interface Props {
  isLoading: boolean;
  isWaitingLoading: boolean;
}

export default function Submit({ isLoading = false, isWaitingLoading = false }: Props) {
  // Form
  const { watch, formState } = useFormContext();
  const { errors } = formState;

  const watchAmount = watch("amount", false);
  const watchAllowance = watch("allowance", false);

  // Wagmi
  const { bridgeEnabled } = useBridge();

  // Context
  const { token, networkLayer } = useChainStore((state) => ({
    token: state.token,
    networkLayer: state.networkLayer,
  }));

  return (
    <div>
      {token && networkLayer !== NetworkLayer.UNKNOWN && token[networkLayer] ? (
        <div className="flex flex-row justify-between">
          <ApproveERC20 />
          <button
            id="submit-erc-btn"
            className={classNames("btn btn-primary w-48 rounded-full uppercase", {
              "cursor-wait": isLoading || isWaitingLoading,
              "btn-disabled": !bridgeEnabled(watchAmount, watchAllowance, errors),
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
          className={classNames("btn w-full btn-primary rounded-full uppercase", {
            "cursor-wait": isLoading || isWaitingLoading,
            "btn-disabled": !bridgeEnabled(watchAmount, BigInt(0), errors),
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
