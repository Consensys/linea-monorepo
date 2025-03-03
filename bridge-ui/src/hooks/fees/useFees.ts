import { useMemo, useCallback } from "react";
import { useAccount } from "wagmi";
import { formatEther, parseUnits, zeroAddress } from "viem";
import { useFormContext } from "react-hook-form";
import useGasFees from "./useGasFees";
import { useChainStore } from "@/stores/chainStore";
import useMinimumFee from "./useMinimumFee";
import { BridgeForm } from "@/models";
import useBridgingFee from "./useBridgingFee";
import useTokenPrices from "../useTokenPrices";
import { useConfigStore } from "@/stores/configStore";

const useFees = () => {
  const { address } = useAccount();
  const fromChain = useChainStore.useFromChain();
  const toChain = useChainStore.useToChain();
  const currency = useConfigStore.useCurrency();
  const { minimumFee } = useMinimumFee();

  const { data: tokenPrices } = useTokenPrices([zeroAddress], fromChain.id);

  const { watch } = useFormContext<BridgeForm>();
  const [claim, token, destinationAddress, amount] = watch(["claim", "token", "destinationAddress", "amount"]);

  const parsedAmount = useMemo(() => parseUnits(amount, token.decimals), [amount, token.decimals]);

  const gasFeesResult = useGasFees({
    address,
    fromChain,
    token,
    recipient: destinationAddress,
    amount: parsedAmount,
    minimumFee,
  });

  const bridgingFees = useBridgingFee({
    account: address,
    token,
    claimingType: claim,
    amount: parsedAmount,
    recipient: destinationAddress,
  });

  const getFiatValue = useCallback(
    (fee: bigint) => {
      const zeroAddrLower = zeroAddress.toLowerCase();
      const price = tokenPrices?.[zeroAddrLower];
      if (!price || price <= 0) return null;
      return Number(formatEther(fee)) * price;
    },
    [tokenPrices],
  );

  const fees = useMemo(() => {
    const feesArray: { name: string; fee: bigint; fiatValue: number | null }[] = [];
    if (gasFeesResult?.gasFees) {
      feesArray.push({
        name: `${fromChain.name} fee`,
        fee: gasFeesResult.gasFees,
        fiatValue: getFiatValue(gasFeesResult.gasFees),
      });
      if (claim === "auto" && bridgingFees) {
        feesArray.push({
          name: `${toChain.name} fee`,
          fee: bridgingFees,
          fiatValue: getFiatValue(bridgingFees),
        });
      }
    }
    return feesArray;
  }, [claim, gasFeesResult?.gasFees, bridgingFees, fromChain.name, toChain.name, tokenPrices, getFiatValue]);

  const totalFees = useMemo(() => {
    const totalFeeBigInt = fees.reduce<bigint>((acc, fee) => acc + fee.fee, 0n);
    const totalFiat = fees.reduce((acc, fee) => acc + Number(fee.fiatValue ?? 0), 0);
    return {
      fees: totalFeeBigInt,
      fiatValue: totalFiat > 0 ? totalFiat : null,
    };
  }, [fees, currency.label]);

  return {
    fees,
    total: totalFees,
  };
};

export default useFees;
