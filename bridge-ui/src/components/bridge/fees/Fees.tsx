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
    bridgingFeeInWei: bigint;
    transactionFeeInWei: bigint;
  };
};

export function Fees({ totalReceived, fees: { total, bridgingFeeInWei, transactionFeeInWei } }: FeesProps) {
  // Context
  const { token, networkLayer, fromChain, networkType } = useChainStore((state) => ({
    token: state.token,
    networkType: state.networkType,
    networkLayer: state.networkLayer,
    fromChain: state.fromChain,
  }));

  // Form
  const {
    watch,
    setError,
    clearErrors,
    setValue,
    formState: { errors },
  } = useFormContext();
  const [amount, claim] = watch(["amount", "claim"]);

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

  useEffect(() => {
    setValue("minFees", bridgingFeeInWei);
  }, [bridgingFeeInWei, setValue]);

  const estimatedTime = networkLayer === NetworkLayer.L1 ? "20 mins" : "8 hrs to 32 hrs";

  return (
    <div className="flex flex-col gap-2 text-sm">
      <FeeLine
        label="Estimated Time"
        value={amount && !errors.amount?.message && estimatedTime}
        tooltip={
          networkLayer === NetworkLayer.L1
            ? "Linea has a 20 minutes delay on deposits as a security measure."
            : "Linea has a minimum 8 hour delay on withdrawals as a security measure. Withdrawals can take up to 32 hours to complete"
        }
      />
      <FeeLine
        label="Estimated Total Fee"
        value={
          isConnected &&
          amount &&
          !errors.amount?.message &&
          (networkType === NetworkType.MAINNET && ethPrice && ethPrice?.[zeroAddress]
            ? `${(Number(formatEther(total)) * ethPrice[zeroAddress].usd).toLocaleString("en-US", {
                style: "currency",
                currency: "USD",
                maximumFractionDigits: 4,
              })}`
            : `${formatBalance(formatEther(total), 8)} ETH`)
        }
        tooltipClassName="before:whitespace-pre-wrap before:content-[attr(data-tip)] text-left"
        tooltip={
          claim === "auto"
            ? `Bridging transaction fee: ${formatBalance(formatEther(transactionFeeInWei), 8)} ETH\nAutomatic claiming Fee: ${formatBalance(formatEther(bridgingFeeInWei), 8)} ETH`
            : `Bridging transaction fee: ${formatBalance(formatEther(transactionFeeInWei), 8)} ETH`
        }
      />
      <FeeLine
        label="Total Received"
        value={
          !errors.amount?.message && totalReceived && totalReceived !== "0"
            ? `${formatBalance(totalReceived)} ${token?.symbol}`
            : undefined
        }
      />
    </div>
  );
}
