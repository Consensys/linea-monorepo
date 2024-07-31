"use client";

import { useMemo, useState } from "react";
import { isAddress, getAddress } from "viem";
import TokenDetails from "./TokenDetails";
import { NetworkType, TokenInfo, TokenType } from "@/config/config";
import fetchTokenInfo from "@/services/fetchTokenInfo";
import useERC20Storage from "@/hooks/useERC20Storage";
import { safeGetAddress } from "@/utils/format";
import { useBridge } from "@/hooks";
import { useChainStore } from "@/stores/chainStore";
import { useTokenStore } from "@/stores/tokenStore";

interface Props {
  tokensModalRef: React.RefObject<HTMLDialogElement>;
}

export default function TokenModal({ tokensModalRef }: Props) {
  const tokensConfig = useTokenStore((state) => state.tokensConfig);
  const [filteredTokens, setFilteredTokens] = useState<TokenInfo[]>([]);
  const [searchTokenIsNew, setSearchTokenIsNew] = useState<boolean>(false);
  const { fillMissingTokenAddress } = useBridge();

  // Context
  const { networkType, networkLayer, fromChain } = useChainStore((state) => ({
    networkType: state.networkType,
    networkLayer: state.networkLayer,
    fromChain: state.fromChain,
  }));
  const { updateOrInsertUserTokenList } = useERC20Storage();
  const [searchQuery, setSearchQuery] = useState("");

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
    <dialog ref={tokensModalRef} id="token_picker_modal" className="modal px-0" onClose={() => setSearchQuery("")}>
      <form method="dialog" className="modal-box overflow-hidden px-0">
        <button id="close-token-picker-modal-btn" className="btn btn-circle btn-ghost btn-sm absolute right-2 top-2">
          ✕
        </button>
        <h3 className="pl-8 text-lg font-bold">Select Token</h3>

        {/* SEARCH FORM */}
        <div className="my-3 flex justify-center border-b border-b-zinc-200 px-7 pb-5 dark:border-b-slate-900/50">
          <input
            type="text"
            placeholder="Search token by name, symbol or address"
            className="input input-bordered w-full"
            onChange={({ target: { value } }) => setSearchQuery(normalizeInput(value))}
            value={searchQuery}
          />
        </div>
        <div className="max-h-[50vh] overflow-auto">
          {filteredTokens.length > 0 ? (
            filteredTokens.map((token: TokenInfo, index: number) => (
              <TokenDetails token={token} onTokenClick={onTokenClick} key={index} />
            ))
          ) : (
            <p className="pl-7 text-error">
              Sorry, there are no results for that term. Please enter a valid token name or address.
            </p>
          )}
        </div>
      </form>
    </dialog>
  );
}
