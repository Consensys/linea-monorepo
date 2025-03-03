import { useFormContext } from "react-hook-form";
import styles from "./received-amount.module.scss";
import { BridgeForm } from "@/models";
import useTokenPrices from "@/hooks/useTokenPrices";
import { useChainStore } from "@/stores/chainStore";
import { useConfigStore } from "@/stores/configStore";

export default function ReceivedAmount() {
  const fromChain = useChainStore.useFromChain();
  const currency = useConfigStore.useCurrency();
  const { watch } = useFormContext<BridgeForm>();

  const [amount, token] = watch(["amount", "token"]);

  const { data: tokenPrices } = useTokenPrices([token[fromChain.layer]], fromChain.id);

  return (
    <div className={styles.value}>
      <p className={styles.crypto}>
        {amount} {token.symbol}
      </p>
      {tokenPrices?.[token[fromChain.layer].toLowerCase()] &&
        tokenPrices?.[token[fromChain.layer].toLowerCase()] > 0 && (
          <p className={styles.amount}>
            {(Number(amount) * tokenPrices?.[token[fromChain.layer].toLowerCase()]).toLocaleString("en-US", {
              style: "currency",
              currency: currency.label,
              maximumFractionDigits: 4,
            })}
          </p>
        )}
    </div>
  );
}
