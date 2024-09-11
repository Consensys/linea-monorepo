"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useFormContext } from "react-hook-form";
import { MdInfoOutline } from "react-icons/md";
import classNames from "classnames";
import { formatEther, parseEther, parseUnits } from "viem";
import { useAccount, useBalance } from "wagmi";
import { useApprove, useExecutionFee } from "@/hooks";
import useBridge from "@/hooks/useBridge";
import { NetworkLayer, TokenType } from "@/config";
import { useChainStore } from "@/stores/chainStore";
import useMinimumFee from "@/hooks/useMinimumFee";

export default function Fees() {
  const [estimatedGasFee, setEstimatedGasFee] = useState<bigint>();

  // Context
  const { tokenBridgeAddress, token, networkLayer, networkType, toChain, fromChain } = useChainStore((state) => ({
    tokenBridgeAddress: state.tokenBridgeAddress,
    token: state.token,
    networkLayer: state.networkLayer,
    networkType: state.networkType,
    toChain: state.toChain,
    fromChain: state.fromChain,
  }));

  // Wagmi
  const { address, isConnected } = useAccount();
  const { data: ethBalance } = useBalance({
    address,
    chainId: fromChain?.id,
  });

  // Form
  const { setValue, register, watch, setError, formState, clearErrors } = useFormContext();
  const { errors } = formState;
  const watchAmount = watch("amount", false);
  const bridgingAllowed = watch("bridgingAllowed", false);
  const claim = watch("claim", false);
  const balance = watch("balance", false);

  // Hooks
  const { minimumFee } = useMinimumFee();
  const { estimateGasBridge } = useBridge();
  const { estimateApprove } = useApprove();
  const minFees = useExecutionFee({
    token,
    claim,
    networkLayer,
    networkType,
    minimumFee,
  });

  // To check which estimateFee and to add the briging fee only if we are bridging
  const enoughAllowance = bridgingAllowed || token?.type === TokenType.ETH;

  useEffect(() => {
    if (minFees) {
      setValue("minFees", minFees);
    }
  }, [minFees, setValue]);

  useEffect(() => {
    const estimate = async () => {
      let calculatedGasFee = BigInt(0);
      if (watchAmount && minimumFee !== null && token?.decimals) {
        if (enoughAllowance) {
          const bridgeGasFee = await estimateGasBridge(watchAmount, minimumFee);
          calculatedGasFee = bridgeGasFee || BigInt(0);
        } else {
          const approveGasFee = await estimateApprove(parseUnits(watchAmount, token.decimals), tokenBridgeAddress);
          calculatedGasFee = approveGasFee || BigInt(0);
        }
      }
      setEstimatedGasFee(calculatedGasFee);
      setValue("gasFees", calculatedGasFee);
    };
    !errors.amount && estimate();
  }, [
    watchAmount,
    minimumFee,
    estimateGasBridge,
    enoughAllowance,
    estimateApprove,
    tokenBridgeAddress,
    token,
    setValue,
    errors.amount,
  ]);

  useEffect(() => {
    if (token?.type === TokenType.ETH && networkLayer === NetworkLayer.L1) {
      setValue("claim", "auto");
    } else {
      setValue("claim", "manual");
    }
  }, [token, setValue, networkLayer]);

  useEffect(() => {
    if (ethBalance && minFees != undefined && parseEther(minFees) > 0 && ethBalance.value <= parseEther(minFees)) {
      setError("minFees", {
        type: "custom",
        message: "Execution fees exceed ETH balance",
      });
    } else {
      clearErrors("minFees");
    }
  }, [setError, balance, minFees, clearErrors, ethBalance]);

  return (
    <ul className="space-y-5">
      <li className="">
        <div className="form-control grid w-full grid-cols-2 gap-6">
          <div>
            <input
              {...register("claim")}
              id="claim-auto"
              type="radio"
              value="auto"
              className="peer hidden"
              disabled={token?.type !== TokenType.ETH}
              required
            />

            <label
              htmlFor="claim-auto"
              className={classNames("btn btn-outline normal-case font-normal w-full", {
                "btn-disabled": networkLayer === NetworkLayer.L2 || token?.type !== TokenType.ETH || !isConnected,

                "peer-checked:btn-outline peer-checked:btn-primary": isConnected,
              })}
            >
              <div className="block">
                <div className="text-md w-full ">Automatic claiming</div>
              </div>
            </label>
          </div>
          <div>
            <input {...register("claim")} id="claim-manual" type="radio" value="manual" className="peer hidden" />
            <label
              htmlFor="claim-manual"
              className={classNames("btn btn-outline normal-case font-normal w-full", {
                "btn-disabled": !isConnected,
                "peer-checked:btn-outline peer-checked:btn-primary": isConnected,
              })}
            >
              <div className="block">
                <div className="text-md w-full">Manual claiming</div>
              </div>
            </label>
          </div>
        </div>
        <div></div>
      </li>
      <li>
        <div
          className={classNames("flex flex-row justify-between", {
            "text-neutral-600": !isConnected,
          })}
        >
          <div className="flex items-center">
            Maximum execution fees :{" "}
            <div
              className="tooltip tooltip-bottom tooltip-info ml-1 "
              data-tip={
                claim === "auto"
                  ? "Automatic bridging: this fee is used to reimburse gas fees paid by the postman to execute the transaction on your behalf on the other chain. If gas fees are lower than the execution fees, the remaining amount  will be reimbursed to the recipient address on the other chain."
                  : "Manual bridging: user will need to claim the funds on the target chain (requires an address funded with ETH for gas fees)"
              }
            >
              <MdInfoOutline className="text-lg" />
            </div>
          </div>
          <span>{minFees} ETH</span>
        </div>

        {errors.minFees && <div className="pt-2 text-right text-error">{errors.minFees.message?.toString()}</div>}
      </li>

      <li
        className={classNames("flex flex-row justify-between", {
          "text-neutral-600": !isConnected,
        })}
      >
        <span>Estimated gas fees:</span>
        {isConnected && <span>{estimatedGasFee ? formatEther(estimatedGasFee) : ""} ETH</span>}
      </li>

      {claim === "manual" && (
        <li
          className={classNames("justify-between", {
            "text-neutral-600": !isConnected,
            "text-white": isConnected,
          })}
        >
          You will have to{" "}
          <Link
            href="https://docs.linea.build/use-mainnet/bridges-of-linea"
            target="_blank"
            referrerPolicy="no-referrer"
            className={classNames("", {
              "text-neutral-600": !isConnected,
              "text-primary": isConnected,
            })}
          >
            claim assets manually
          </Link>{" "}
          once the transaction reaches the other layer.{" "}
          {networkLayer === NetworkLayer.L1
            ? `This can take up to ~20min. You will need ETH on ${toChain?.name} to pay for gas fees.`
            : `This can take between 8 and 32 hours. You will need ETH on ${toChain?.name} to pay for gas fees.`}
        </li>
      )}
    </ul>
  );
}
