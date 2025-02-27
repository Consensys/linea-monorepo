"use client";

import React, { useEffect } from "react";
import Image from "next/image";
import { FieldValues, UseFormClearErrors, UseFormSetValue } from "react-hook-form";
import { useBlockNumber } from "wagmi";
import { formatBalance } from "@/utils/format";
import { TokenInfo, TokenType } from "@/config/config";
import { useTokenBalance } from "@/hooks/useTokenBalance";
import styles from "./token-details.module.scss";
import { useTokenStore } from "@/stores/tokenStoreProvider";
import { useChainStore } from "@/stores/chainStore";
import { CurrencyOption } from "@/stores/configStore";

interface TokenDetailsProps {
  token: TokenInfo;
  onTokenClick: (token: TokenInfo) => void;
  setValue: UseFormSetValue<FieldValues>;
  clearErrors: UseFormClearErrors<FieldValues>;
  tokenPrice?: number;
  currency: CurrencyOption;
}

export default function TokenDetails({
  token,
  onTokenClick,
  setValue,
  clearErrors,
  tokenPrice,
  currency,
}: TokenDetailsProps) {
  const setSelectedToken = useTokenStore((state) => state.setSelectedToken);
  const fromChain = useChainStore.useFromChain();

  const tokenNotFromCurrentLayer = fromChain?.layer && !token[fromChain?.layer] && token.type !== TokenType.ETH;

  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { balance, refetch } = useTokenBalance(token[fromChain.layer], token?.decimals);

  useEffect(() => {
    if (blockNumber && blockNumber % 5n === 0n) {
      refetch();
    }
  }, [blockNumber, refetch]);

  return (
    <button
      id={`token-details-${token.symbol}-btn`}
      className={styles["token-wrapper"]}
      type="button"
      disabled={tokenNotFromCurrentLayer}
      onClick={() => {
        setValue("amount", "");
        clearErrors("amount");
        setSelectedToken(token);
        setValue("token", token);
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
            {formatBalance(balance)} {token.symbol}
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
