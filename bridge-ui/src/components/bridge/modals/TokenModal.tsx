"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { isAddress, getAddress, Address, zeroAddress } from "viem";
import TokenDetails from "./TokenDetails";
import { NetworkType, TokenInfo, TokenType } from "@/config/config";
import { useERC20Storage, useTokenFetch } from "@/hooks";
import { safeGetAddress } from "@/utils/format";
import { useChainStore } from "@/stores/chainStore";
import { useTokenStore } from "@/stores/tokenStoreProvider";
import { fetchTokenInfo } from "@/services";
import { FieldValues, UseFormClearErrors, UseFormSetValue } from "react-hook-form";
import { CiSearch } from "react-icons/ci";
import useTokenPrices from "@/hooks/useTokenPrices";
import { isEmptyObject } from "@/utils/utils";
import useDebounce from "@/hooks/useDebounce";

interface TokenModalProps {
  setValue: UseFormSetValue<FieldValues>;
  clearErrors: UseFormClearErrors<FieldValues>;
}

export default function TokenModal({ setValue, clearErrors }: TokenModalProps) {
  const tokensList = useTokenStore((state) => state.tokensList);
  const [filteredTokens, setFilteredTokens] = useState<TokenInfo[]>([]);
  const [searchTokenIsNew, setSearchTokenIsNew] = useState<boolean>(false);
  const { fillMissingTokenAddress } = useTokenFetch();

  // Context
  const networkType = useChainStore((state) => state.networkType);
  const networkLayer = useChainStore((state) => state.networkLayer);
  const fromChain = useChainStore((state) => state.fromChain);

  const { updateOrInsertUserTokenList } = useERC20Storage();
  const [searchQuery, setSearchQuery] = useState("");

  const { data } = useTokenPrices(
    useMemo(
      () =>
        filteredTokens.map((token) =>
          token.name === "Ether" ? zeroAddress : (safeGetAddress(token[networkLayer]) as Address),
        ),
      [filteredTokens, networkLayer],
    ),
    fromChain?.id,
  );

  const debouncedSearchQuery = useDebounce(searchQuery, 300);

  const handleTokenSearch = useCallback(
    async (query: string) => {
      let found = false;

      if (networkType === NetworkType.SEPOLIA || networkType === NetworkType.MAINNET) {
        const currentNetworkTokens = tokensList?.[networkType] || [];

        // Filter tokens based on the search query
        const filtered = currentNetworkTokens.filter(
          (token: TokenInfo) =>
            (token[networkLayer] || token.type === TokenType.ETH) &&
            (token.name.toLowerCase().includes(query) ||
              token.symbol.toLowerCase().includes(query) ||
              safeGetAddress(token[networkLayer])?.toLowerCase().includes(query)),
        );

        if (filtered.length > 0) {
          found = true;
          setFilteredTokens(filtered);
          setSearchTokenIsNew(false);
        } else if (isAddress(query)) {
          // Fetch token info from the contract if the query is a valid address
          const newToken = await fetchTokenInfo(query, networkType, fromChain);
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
    },
    [networkType, networkLayer, tokensList, fromChain, fillMissingTokenAddress],
  );

  const handleTokenClick = useCallback(
    (token: TokenInfo) => {
      if (searchTokenIsNew && token[networkLayer]) {
        updateOrInsertUserTokenList(token, networkType);
      }
      setSearchTokenIsNew(false);
    },
    [searchTokenIsNew, networkLayer, updateOrInsertUserTokenList, networkType],
  );

  const getTokenPrice = useCallback(
    (token: TokenInfo): number | undefined => {
      if (networkType === NetworkType.MAINNET && !isEmptyObject(data)) {
        const tokenAddress = (safeGetAddress(token[networkLayer]) || zeroAddress).toLowerCase();
        return data[tokenAddress]?.usd;
      }
      return undefined;
    },
    [networkType, networkLayer, data],
  );

  const normalizeInput = (input: string): string => {
    return isAddress(input) ? getAddress(input) : input.toLowerCase();
  };

  useEffect(() => {
    if (debouncedSearchQuery.trim() === "") {
      if (networkType === NetworkType.SEPOLIA || networkType === NetworkType.MAINNET) {
        setFilteredTokens(tokensList[networkType]);
        setSearchTokenIsNew(false);
        return;
      }
      setFilteredTokens([]);
      setSearchTokenIsNew(false);
      return;
    }
    handleTokenSearch(debouncedSearchQuery);
  }, [debouncedSearchQuery, handleTokenSearch, networkType, tokensList]);

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
                onTokenClick={handleTokenClick}
                key={`token-details-${index}`}
                setValue={setValue}
                clearErrors={clearErrors}
                tokenPrice={getTokenPrice(token)}
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
