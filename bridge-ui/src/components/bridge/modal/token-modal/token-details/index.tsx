"use client";

import React from "react";
import Image from "next/image";
import { TokenInfo } from "@/config/config";
import styles from "./token-details.module.scss";
import { useTokenStore } from "@/stores/tokenStoreProvider";
import { useChainStore } from "@/stores/chainStore";
import { CurrencyOption } from "@/stores/configStore";
import { isEth } from "@/utils/tokens";
import { useTokenBalance } from "@/hooks/useTokenBalance";
import { useFormStore } from "@/stores/formStoreProvider";
import { formatUnits } from "viem";

interface TokenDetailsProps {
  token: TokenInfo;
  onTokenClick: (token: TokenInfo) => void;
  tokenPrice?: number;
  currency: CurrencyOption;
}

export default function TokenDetails({ token, onTokenClick, tokenPrice, currency }: TokenDetailsProps) {
  const setSelectedToken = useTokenStore((state) => state.setSelectedToken);
  const fromChain = useChainStore.useFromChain();
  const { balance } = useTokenBalance(token);
  const setToken = useFormStore((state) => state.setToken);
  const setAmount = useFormStore((state) => state.setAmount);

  const tokenNotFromCurrentLayer = fromChain?.layer && !token[fromChain?.layer] && !isEth(token);

  return (
    <button
      id={`token-details-${token.symbol}-btn`}
      className={styles["token-wrapper"]}
      type="button"
      disabled={tokenNotFromCurrentLayer}
      onClick={() => {
        setAmount(0n);
        setSelectedToken(token);
        setToken(token);
        onTokenClick(token);
      }}
    >
      <div className={styles["left"]}>
        <Image src={token.image} alt={token.name} width={32} height={32} />
        <div className={styles["text-left"]}>
          <p className={styles["token-symbol"]}>{token.symbol}</p>
          <p className={styles["token-name"]}>{token.name}</p>
        </div>
      </div>
      {!tokenNotFromCurrentLayer && (
        <div className={styles.rìght}>
          <p className={styles["balance"]}>
            {formatUnits(balance, token.decimals)} {token.symbol}
          </p>
          {tokenPrice ? (
            <p className={styles["price"]}>
              {(tokenPrice * Number(balance)).toLocaleString("en-US", {
                style: "currency",
                currency: currency.label,
                maximumFractionDigits: 4,
              })}
            </p>
          ) : null}
        </div>
      )}
      {tokenNotFromCurrentLayer && (
        <div className={styles["not-from-current-layer"]}>
          <p>Token is from other layer. Please swap networks to import token.</p>
        </div>
      )}
    </button>
  );
}
