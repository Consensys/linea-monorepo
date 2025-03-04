import { useMemo, useCallback } from "react";
import { useAccount } from "wagmi";
import { formatEther, zeroAddress } from "viem";
import useGasFees from "./useGasFees";
import { useChainStore } from "@/stores/chainStore";
import useMinimumFee from "./useMinimumFee";
import useBridgingFee from "./useBridgingFee";
import useTokenPrices from "../useTokenPrices";
import { useFormStore } from "@/stores/formStoreProvider";

const useFees = () => {
  const { address } = useAccount();
  const fromChain = useChainStore.useFromChain();
  const toChain = useChainStore.useToChain();
  const { minimumFee } = useMinimumFee();

  const { data: tokenPrices } = useTokenPrices([zeroAddress], fromChain.id);

  const claim = useFormStore((state) => state.claim);
  const token = useFormStore((state) => state.token);
  const recipient = useFormStore((state) => state.recipient);
  const amount = useFormStore((state) => state.amount);

  const gasFeesResult = useGasFees({
    address,
    fromChain,
    token,
    recipient,
    amount: amount ?? 0n,
    minimumFee,
  });

  const bridgingFees = useBridgingFee({
    account: address,
    token,
    claimingType: claim,
    amount: amount ?? 0n,
    recipient,
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
  }, [gasFeesResult?.gasFees, fromChain.name, getFiatValue, claim, bridgingFees, toChain.name]);

  const totalFees = useMemo(() => {
    const totalFeeBigInt = fees.reduce<bigint>((acc, fee) => acc + fee.fee, 0n);
    const totalFiat = fees.reduce((acc, fee) => acc + Number(fee.fiatValue ?? 0), 0);
    return {
      fees: totalFeeBigInt,
      fiatValue: totalFiat > 0 ? totalFiat : null,
    };
  }, [fees]);

  return {
    fees,
    total: totalFees,
  };
};

export default useFees;
