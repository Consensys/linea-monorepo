"use client";

import React, { useEffect } from "react";
import Image from "next/image";
import { useFormContext } from "react-hook-form";
import { useAccount, useBalance, useBlockNumber } from "wagmi";
import classNames from "classnames";

import { config } from "@/config";
import { formatBalance } from "@/utils/format";
import { NetworkLayer, NetworkType, TokenInfo, TokenType } from "@/config/config";
import { useQueryClient } from "@tanstack/react-query";
import { useChainStore } from "@/stores/chainStore";

interface TokenDetailsProps {
  token: TokenInfo;
  onTokenClick: (token: TokenInfo) => void;
}
export default function TokenDetails({ token, onTokenClick }: TokenDetailsProps) {
  const { address } = useAccount();
  const { networkLayer, fromChain, setToken, setTokenBridgeAddress, networkType } = useChainStore((state) => ({
    networkLayer: state.networkLayer,
    fromChain: state.fromChain,
    setToken: state.setToken,
    setTokenBridgeAddress: state.setTokenBridgeAddress,
    networkType: state.networkType,
  }));

  const tokenNotFromCurrentLayer = !token[networkLayer] && token.type !== TokenType.ETH;

  // Form
  const { setValue, clearErrors } = useFormContext();

  const queryClient = useQueryClient();
  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { data: balance, queryKey } = useBalance({
    address,
    token: token[networkLayer] ?? undefined,
    chainId: fromChain?.id,
  });

  useEffect(() => {
    if (blockNumber && blockNumber % 5n === 0n) {
      queryClient.invalidateQueries({ queryKey });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [blockNumber, queryClient]);

  return (
    <button
      id={`token-details-${token.symbol}-btn`}
      className={classNames(
        "flex items-center justify-between w-full gap-5 px-8 py-3 mt-3 bg-transparent border-0 hover:bg-slate-900/20",
        {
          "btn-disabled": tokenNotFromCurrentLayer,
        },
      )}
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
      <div className="flex gap-5">
        <Image
          src={token.image}
          alt={token.name}
          width={0}
          height={0}
          style={{ width: "40px", height: "auto" }}
          className="rounded-full"
        />
        <div className="text-left">
          <p className="text-semibold">{token.name}</p>
          <p className="text-sm text-zinc-300">{token.symbol}</p>
        </div>
      </div>
      {!tokenNotFromCurrentLayer && (
        <div className="text-right">
          <p>Balance</p>
          <p className="text-sm text-zinc-300">
            {formatBalance(balance?.formatted)} {balance?.symbol}
          </p>
        </div>
      )}
      {tokenNotFromCurrentLayer && (
        <div className="text-left text-warning">
          <p>Token is from other layer. Please swap networks to import token.</p>
        </div>
      )}
    </button>
  );
}
