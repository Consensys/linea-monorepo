import { formatUnits } from "viem";
import styles from "./received-amount.module.scss";
import { useTokenPrices } from "@/hooks";
import { useConfigStore, useChainStore, useFormStore } from "@/stores";
import { formatBalance } from "@/utils";
import { ETH_SYMBOL } from "@/constants";
import { ChainLayer } from "@/types";
import { useMemo } from "react";

function formatReceivedAmount(
  amount: bigint,
  tokenSymbol: string,
  bridgingFees: bigint,
  minimumFees: bigint,
  fromChainLayer: ChainLayer,
) {
  if (tokenSymbol !== ETH_SYMBOL) {
    return amount;
  }

  return fromChainLayer === ChainLayer.L1 ? amount - bridgingFees : amount - minimumFees;
}

export default function ReceivedAmount() {
  const { fromChainId, fromChainLayer } = useChainStore((state) => ({
    fromChainId: state.fromChain.id,
    fromChainLayer: state.fromChain.layer,
  }));
  const currency = useConfigStore((state) => state.currency);
  const amount = useFormStore((state) => state.amount);
  const token = useFormStore((state) => state.token);
  const bridgingFees = useFormStore((state) => state.bridgingFees);
  const minimumFees = useFormStore((state) => state.minimumFees);

  const { data: tokenPrices } = useTokenPrices([token[fromChainLayer]], fromChainId);

  const receivedAmount = useMemo(
    () =>
      formatUnits(
        formatReceivedAmount(amount || 0n, token.symbol, bridgingFees, minimumFees, fromChainLayer),
        token.decimals,
      ),
    [amount, token.symbol, bridgingFees, minimumFees, fromChainLayer, token.decimals],
  );

  return (
    <div className={styles.value}>
      <p className={styles.crypto} data-testid="received-amount-text">
        {formatBalance(receivedAmount, 6)} {token.symbol}
      </p>
      {tokenPrices?.[token[fromChainLayer].toLowerCase()] && tokenPrices?.[token[fromChainLayer].toLowerCase()] > 0 && (
        <p className={styles.amount}>
          {(Number(receivedAmount) * tokenPrices?.[token[fromChainLayer].toLowerCase()]).toLocaleString("en-US", {
            style: "currency",
            currency: currency.label,
            maximumFractionDigits: 8,
          })}
        </p>
      )}
    </div>
  );
}
