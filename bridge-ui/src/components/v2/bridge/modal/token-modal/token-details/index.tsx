"use client";

import React, { useEffect } from "react";
import Image from "next/image";
import { FieldValues, UseFormClearErrors, UseFormSetValue } from "react-hook-form";
import { useBlockNumber } from "wagmi";
import { config } from "@/config";
import { formatBalance } from "@/utils/format";
import { NetworkLayer, NetworkType, TokenInfo, TokenType } from "@/config/config";
import { useChainStore } from "@/stores/chainStore";
import { useTokenBalance } from "@/hooks/useTokenBalance";
import styles from "./token-details.module.scss";

interface TokenDetailsProps {
  token: TokenInfo;
  onTokenClick: (token: TokenInfo) => void;
  setValue: UseFormSetValue<FieldValues>;
  clearErrors: UseFormClearErrors<FieldValues>;
  tokenPrice?: number;
}

export default function TokenDetails({ token, onTokenClick, setValue, clearErrors, tokenPrice }: TokenDetailsProps) {
  const { networkLayer, setToken, setTokenBridgeAddress, networkType } = useChainStore((state) => ({
    networkLayer: state.networkLayer,
    setToken: state.setToken,
    setTokenBridgeAddress: state.setTokenBridgeAddress,
    networkType: state.networkType,
  }));

  const tokenNotFromCurrentLayer = !token[networkLayer] && token.type !== TokenType.ETH;

  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { balance, refetch } = useTokenBalance(token[networkLayer], token?.decimals);

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
        if (networkLayer !== NetworkLayer.UNKNOWN && token && networkType !== NetworkType.WRONG_NETWORK) {
          setValue("amount", "");
          clearErrors("amount");
          setToken(token);
          switch (token.type) {
            case TokenType.USDC:
              setTokenBridgeAddress(config.networks[networkType][networkLayer].usdcBridgeAddress);
              break;
            default:
              setTokenBridgeAddress(config.networks[networkType][networkLayer].tokenBridgeAddress);
              break;
          }
          onTokenClick(token);
        }
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
                currency: "USD",
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
