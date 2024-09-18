"use client";

import { useMemo, useState } from "react";
import { isAddress, getAddress, Address, zeroAddress } from "viem";
import TokenDetails from "./TokenDetails";
import { NetworkType, TokenInfo, TokenType } from "@/config/config";
import { useERC20Storage, useTokenFetch } from "@/hooks";
import { safeGetAddress } from "@/utils/format";
import { useChainStore } from "@/stores/chainStore";
import { useTokenStore } from "@/stores/tokenStore";
import { fetchTokenInfo } from "@/services";
import { FieldValues, UseFormClearErrors, UseFormSetValue } from "react-hook-form";
import { CiSearch } from "react-icons/ci";
import useTokenPrices from "@/hooks/useTokenPrices";
import { isEmptyObject } from "@/utils/utils";

interface TokenModalProps {
  setValue: UseFormSetValue<FieldValues>;
  clearErrors: UseFormClearErrors<FieldValues>;
}

export default function TokenModal({ setValue, clearErrors }: TokenModalProps) {
  const tokensConfig = useTokenStore((state) => state.tokensConfig);
  const [filteredTokens, setFilteredTokens] = useState<TokenInfo[]>([]);
  const [searchTokenIsNew, setSearchTokenIsNew] = useState<boolean>(false);
  const { fillMissingTokenAddress } = useTokenFetch();

  // Context
  const { networkType, networkLayer, fromChain } = useChainStore((state) => ({
    networkType: state.networkType,
    networkLayer: state.networkLayer,
    fromChain: state.fromChain,
  }));
  const { updateOrInsertUserTokenList } = useERC20Storage();
  const [searchQuery, setSearchQuery] = useState("");
  const { data } = useTokenPrices(
    filteredTokens.map((token) =>
      token.name === "Ether" ? zeroAddress : (safeGetAddress(token[networkLayer]) as Address),
    ),
    fromChain?.id,
  );

  useMemo(async () => {
    let found = false;
    if (networkType === NetworkType.SEPOLIA || networkType === NetworkType.MAINNET) {
      const filtered = (tokensConfig?.[networkType] ?? []).filter(
        (token: TokenInfo) =>
          (token[networkLayer] || token.type === TokenType.ETH) &&
          (token.name.toLowerCase()?.includes(searchQuery) ||
            token.symbol.toLowerCase()?.includes(searchQuery) ||
            safeGetAddress(token[networkLayer])?.includes(searchQuery)),
      );

      if (filtered.length > 0) {
        found = true;
        setFilteredTokens(filtered);
        setSearchTokenIsNew(false);
      } else if (isAddress(searchQuery)) {
        // Get token info from contract
        const newToken = await fetchTokenInfo(searchQuery, networkType, fromChain);
        if (newToken) {
          await fillMissingTokenAddress(newToken);
          found = true;
          setFilteredTokens([newToken]);
          setSearchTokenIsNew(true);
        } else {
          setSearchTokenIsNew(false);
        }
      } else {
        setSearchTokenIsNew(false);
      }
    }
    if (!found) {
      setFilteredTokens([]);
    }
  }, [searchQuery, networkType, networkLayer, tokensConfig, fromChain, fillMissingTokenAddress]);

  const onTokenClick = (token: TokenInfo) => {
    if (searchTokenIsNew && token[networkLayer]) {
      updateOrInsertUserTokenList(token, networkType);
    }

    setSearchTokenIsNew(false);
  };

  const normalizeInput = (input: string): string => {
    if (isAddress(input)) {
      return getAddress(input);
    } else {
      return input.toLowerCase();
    }
  };

  return (
    <div id="token-picker-modal">
      <form method="dialog" className="overflow-hidden">
        <div className="mt-1 flex justify-center px-4">
          <label className="input input-bordered flex w-full items-center gap-2 rounded-full bg-inherit focus:border-0 focus:outline-none">
            <CiSearch size="20" />
            <input
              type="text"
              className="grow"
              placeholder="Search token by name, symbol or address"
              onChange={({ target: { value } }) => setSearchQuery(normalizeInput(value))}
              value={searchQuery}
            />
          </label>
        </div>
        <div className="mt-3 max-h-[50vh] overflow-auto">
          {filteredTokens.length > 0 ? (
            filteredTokens.map((token: TokenInfo, index: number) => (
              <TokenDetails
                token={token}
                onTokenClick={onTokenClick}
                key={index}
                setValue={setValue}
                clearErrors={clearErrors}
                tokenPrice={
                  (networkType === NetworkType.MAINNET &&
                    !isEmptyObject(data) &&
                    data[safeGetAddress(token[networkLayer])?.toLowerCase() || zeroAddress]?.usd) ||
                  undefined
                }
              />
            ))
          ) : (
            <p className="pl-7 text-error">
              Sorry, there are no results for that term. Please enter a valid token name or address.
            </p>
          )}
        </div>
      </form>
    </div>
  );
}
