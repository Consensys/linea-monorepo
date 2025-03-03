import { ChangeEvent, useEffect } from "react";
import { useAccount } from "wagmi";
import { useFormContext } from "react-hook-form";
import { useIsLoggedIn } from "@/lib/dynamic";
import useTokenPrices from "@/hooks/useTokenPrices";
import { useChainStore } from "@/stores/chainStore";
import styles from "./amount.module.scss";
import { useConfigStore } from "@/stores/configStore";

const AMOUNT_REGEX = "^[0-9]*[.,]?[0-9]*$";
const MAX_AMOUNT_CHAR = 20;

export function Amount() {
  const currency = useConfigStore((state) => state.currency);
  const fromChain = useChainStore.useFromChain();

  const { address } = useAccount();
  const isLoggedIn = useIsLoggedIn();

  const { setValue, getValues, trigger } = useFormContext();
  const [amount, token] = getValues(["amount", "token"]);
  const tokenAddress = token[fromChain.layer];

  const { data: tokenPrices } = useTokenPrices([tokenAddress], fromChain.id);

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    const { key } = event;

    // Allow control keys, numeric keys, decimal point (if not already present), +, -, and arrow keys
    const allowedKeys = ["Backspace", "Tab", "ArrowLeft", "ArrowRight", "Delete"];

    if (/[0-9]/.test(key) && !amount.includes(".") && amount[0] === "0") {
      event.preventDefault();
      return;
    }
    if (
      !(
        /[0-9]/.test(key) ||
        allowedKeys.includes(key) ||
        (key === "." && !amount.includes(".")) ||
        (key === "," && !amount.includes(","))
      )
    ) {
      event.preventDefault();
    }
  };

  const checkAmountHandler = (e: ChangeEvent<HTMLInputElement>) => {
    // Replace minus
    const amount = e.target.value.replace(/,/g, ".");

    if (!token) {
      return;
    }

    if (new RegExp(AMOUNT_REGEX).test(amount) || amount === "") {
      // Limit max char
      if (amount.length > MAX_AMOUNT_CHAR) {
        setValue("amount", amount.substring(0, MAX_AMOUNT_CHAR));
      } else {
        setValue("amount", amount);
      }
    }
  };

  useEffect(() => {
    if (amount) {
      trigger(["amount"]);
    }
  }, [amount, trigger]);

  useEffect(() => {
    setValue("amount", "");
  }, [address, setValue]);

  return (
    <div className={styles["amount"]}>
      <p className={styles.title}>Send</p>
      <input
        disabled={!isLoggedIn}
        id="amount-input"
        type="text"
        autoCorrect="off"
        autoComplete="off"
        spellCheck="false"
        inputMode="decimal"
        value={amount}
        onKeyDown={handleKeyDown}
        onChange={checkAmountHandler}
        pattern={AMOUNT_REGEX}
        placeholder="0"
      />
      {!fromChain?.testnet && (
        <span className={styles["calculated-value"]}>
          {amount && tokenPrices?.[tokenAddress.toLowerCase()] && tokenPrices?.[tokenAddress.toLowerCase()] > 0 ? (
            <>
              {(Number(amount) * tokenPrices?.[tokenAddress.toLowerCase()]).toLocaleString("en-US", {
                style: "currency",
                currency: currency.label,
                maximumFractionDigits: 4,
              })}
            </>
          ) : (
            ""
          )}
        </span>
      )}
    </div>
  );
}
