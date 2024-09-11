import { useEffect } from "react";
import { useAccount, useBalance } from "wagmi";
import { formatEther, zeroAddress } from "viem";
import { useFormContext } from "react-hook-form";
import { NetworkLayer, NetworkType } from "@/config";
import { useChainStore } from "@/stores/chainStore";
import { formatBalance } from "@/utils/format";
import { FeeLine } from "./FeeLine";
import useTokenPrices from "@/hooks/useTokenPrices";

type FeesProps = {
  totalReceived: string;
  fees: {
    total: bigint;
    executionFeeInWei: bigint;
    bridgingFeeInWei: bigint;
  };
};

export function Fees({ totalReceived, fees: { total, executionFeeInWei, bridgingFeeInWei } }: FeesProps) {
  // Context
  const { token, networkLayer, fromChain, networkType } = useChainStore((state) => ({
    token: state.token,
    networkType: state.networkType,
    networkLayer: state.networkLayer,
    fromChain: state.fromChain,
  }));

  // Form
  const { watch, setError, clearErrors } = useFormContext();
  const amount = watch("amount");

  // Wagmi
  const { address, isConnected } = useAccount();
  const { data: ethBalance } = useBalance({
    address,
    chainId: fromChain?.id,
  });

  // Hooks
  const { data: ethPrice } = useTokenPrices([zeroAddress], fromChain?.id);

  useEffect(() => {
    if (ethBalance && total && total > 0 && ethBalance.value <= total) {
      setError("minFees", {
        type: "custom",
        message: "Execution fees exceed ETH balance",
      });
    } else {
      clearErrors("minFees");
    }
  }, [setError, clearErrors, ethBalance, total]);

  const estimatedTime = networkLayer === NetworkLayer.L1 ? "20 mins" : "8 hrs to 32 hrs";

  return (
    <div className="flex flex-col gap-2 text-sm">
      <FeeLine
        label="Estimated Time"
        value={amount && estimatedTime}
        tooltip="Linea has a minimum 8 hour delay on withdrawals as a security measure. 
Withdrawals can take up to 32 hours to complete"
      />
      <FeeLine
        label="Estimated Total Fee"
        value={
          isConnected &&
          amount &&
          (networkType === NetworkType.MAINNET && ethPrice && ethPrice?.[zeroAddress]
            ? `$${(Number(formatEther(total)) * ethPrice[zeroAddress].usd).toLocaleString("en-US", {
                minimumFractionDigits: 2,
                maximumFractionDigits: 10,
              })}`
            : `${formatBalance(formatEther(total))} ETH`)
        }
        tooltipClassName="before:whitespace-pre-wrap before:content-[attr(data-tip)] text-left"
        tooltip={`Execution Fee: ${formatEther(executionFeeInWei)} ETH\nBridging fee: ${formatEther(bridgingFeeInWei)} ETH`}
      />
      <FeeLine
        label="Total Received"
        value={totalReceived && totalReceived !== "0" ? `${formatBalance(totalReceived)} ${token?.symbol}` : undefined}
      />
    </div>
  );
}
