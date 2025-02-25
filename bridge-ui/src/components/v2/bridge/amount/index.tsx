import { ChangeEvent, useCallback, useEffect } from "react";
import { useAccount } from "wagmi";
import { useFormContext } from "react-hook-form";
import { parseUnits, zeroAddress } from "viem";

import { NetworkType, TokenType } from "@/config";
import useTokenPrices from "@/hooks/useTokenPrices";
import { useChainStore } from "@/stores/chainStore";

import styles from "./amount.module.scss";

const AMOUNT_REGEX = "^[0-9]*[.,]?[0-9]*$";
const MAX_AMOUNT_CHAR = 20;

export function Amount() {
  const token = useChainStore.useToken();
  const fromChain = useChainStore.useFromChain();
  const networkLayer = useChainStore.useNetworkLayer();
  const networkType = useChainStore.useNetworkType();
  const tokenAddress = token?.[networkLayer] || zeroAddress;

  const { address } = useAccount();

  const { setValue, getValues, setError, clearErrors, trigger, watch } = useFormContext();
  const watchBalance = watch("balance");
  const [amount, gasFees, minFees] = getValues(["amount", "gasFees", "minFees"]);

  const { data: tokenPrices } = useTokenPrices([tokenAddress], fromChain?.id);

  const compareAmountBalance = useCallback(
    (_amount: string) => {
      if (!token) {
        return;
      }
      const amountToCompare =
        token.type === TokenType.ETH
          ? parseUnits(_amount, token.decimals) + gasFees + minFees
          : parseUnits(_amount, token.decimals);

      const balanceToCompare = parseUnits(watchBalance, token.decimals);

      if (amountToCompare > balanceToCompare) {
        setError("amount", {
          type: "custom",
          message: "Not enough funds (Incl fees)",
        });
      } else {
        clearErrors("amount");
      }
    },
    [token, gasFees, minFees, clearErrors, setError, watchBalance],
  );

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
    compareAmountBalance(amount);
  };

  useEffect(() => {
    if (amount) {
      trigger(["amount"]);
      compareAmountBalance(amount);
    }
  }, [amount, trigger, compareAmountBalance]);

  // Detect when changing account
  useEffect(() => {
    setValue("amount", "");
    clearErrors("amount");
  }, [address, setValue, clearErrors]);

  return (
    <div className={styles["amount"]}>
      <p className={styles.title}>Send</p>
      <input
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
      {networkType === NetworkType.MAINNET && (
        <span className={styles["calculated-value"]}>
          {amount &&
          tokenPrices?.[tokenAddress.toLowerCase()]?.usd &&
          tokenPrices?.[tokenAddress.toLowerCase()]?.usd > 0 ? (
            <>
              {(Number(amount) * tokenPrices?.[tokenAddress.toLowerCase()]?.usd).toLocaleString("en-US", {
                style: "currency",
                currency: "USD",
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
