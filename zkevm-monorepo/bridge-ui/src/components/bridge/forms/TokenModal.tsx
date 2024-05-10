'use client';

import { useContext, useMemo, useState } from 'react';
import { isAddress, getAddress } from 'viem';

import TokenDetails from './TokenDetails';
import { ChainContext, NetworkType } from '@/contexts/chain.context';
import { TokenInfo, TokenType } from '@/config/config';
import fetchTokenInfo from '@/services/fetchTokenInfo';
import useERC20Storage from '@/hooks/useERC20Storage';
import { safeGetAddress } from '@/utils/format';
import { useConfigContext } from '@/contexts/token.context';
import { useBridge } from '@/hooks';

interface Props {
  tokensModalRef: React.RefObject<HTMLDialogElement>;
}

export default function TokenModal({ tokensModalRef }: Props) {
  const { tokensConfig } = useConfigContext();
  const [filteredTokens, setFilteredTokens] = useState<TokenInfo[]>([]);
  const [searchTokenIsNew, setSearchTokenIsNew] = useState<boolean>(false);
  const { fillMissingTokenAddress } = useBridge();

  // Context
  const { networkType, networkLayer, activeChain } = useContext(ChainContext);
  const { updateOrInsertUserTokenList } = useERC20Storage();
  const [searchQuery, setSearchQuery] = useState('');

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
        const newToken = await fetchTokenInfo(searchQuery, networkType, activeChain);
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
  }, [searchQuery, networkType, networkLayer, tokensConfig, activeChain, fillMissingTokenAddress]);

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
    <dialog ref={tokensModalRef} id="token_picker_modal" className="px-0 modal" onClose={() => setSearchQuery('')}>
      <form method="dialog" className="px-0 overflow-hidden modal-box">
        <button id="close-token-picker-modal-btn" className="absolute btn btn-sm btn-circle btn-ghost right-2 top-2">
          ✕
        </button>
        <h3 className="pl-8 text-lg font-bold">Select Token</h3>

        {/* SEARCH FORM */}
        <div className="flex justify-center pb-5 my-3 border-b px-7 border-b-zinc-200 dark:border-b-slate-900/50">
          <input
            type="text"
            placeholder="Search token by name, symbol or address"
            className="w-full input input-bordered"
            onChange={({ target: { value } }) => setSearchQuery(normalizeInput(value))}
            value={searchQuery}
          />
        </div>
        <div className="overflow-auto max-h-[50vh]">
          {filteredTokens.length > 0 ? (
            filteredTokens.map((token: TokenInfo, index: number) => (
              <TokenDetails token={token} onTokenClick={onTokenClick} key={index} />
            ))
          ) : (
            <p className="text-error pl-7">
              Sorry, there are no results for that term. Please enter a valid token name or address.
            </p>
          )}
        </div>
      </form>
    </dialog>
  );
}
