"use client";

import { useCallback, useMemo, useState } from "react";
import { isAddress, getAddress, Address, zeroAddress } from "viem";
import { TokenInfo } from "@/config/config";
import { safeGetAddress } from "@/utils/format";
import { useChainStore } from "@/stores/chainStore";
import useTokenPrices from "@/hooks/useTokenPrices";
import { isEmptyObject } from "@/utils/utils";
import Modal from "@/components/modal";
import SearchIcon from "@/assets/icons/search.svg";
import styles from "./token-modal.module.scss";
import TokenDetails from "./token-details";
import { useDevice } from "@/hooks/useDevice";
import { useTokens } from "@/hooks/useTokens";
import { useTokenStore } from "@/stores/tokenStoreProvider";
import { useConfigStore } from "@/stores/configStore";
import { isEth } from "@/utils/tokens";

interface TokenModalProps {
  isModalOpen: boolean;
  onCloseModal: () => void;
}

export default function TokenModal({ isModalOpen, onCloseModal }: TokenModalProps) {
  const tokensList = useTokens();
  const setSelectedToken = useTokenStore((state) => state.setSelectedToken);
  const fromChain = useChainStore.useFromChain();
  const currency = useConfigStore((state) => state.currency);
  const { isMobile } = useDevice();

  const [searchQuery, setSearchQuery] = useState("");

  const filteredTokens = useMemo(() => {
    if (!searchQuery) return tokensList;
    const query = searchQuery.toLowerCase();

    return tokensList.filter((token: TokenInfo) => {
      const tokenAddress = fromChain?.layer ? token[fromChain.layer] : undefined;
      return (
        (tokenAddress || isEth(token)) &&
        (token.name.toLowerCase().includes(query) ||
          token.symbol.toLowerCase().includes(query) ||
          (tokenAddress && safeGetAddress(tokenAddress)?.toLowerCase().includes(query)))
      );
    });
  }, [tokensList, searchQuery, fromChain?.layer]);

  const tokenAddresses = useMemo(
    () =>
      filteredTokens.map((token) => {
        if (token.name === "Ether") return zeroAddress;
        const tokenAddress = fromChain?.layer ? safeGetAddress(token[fromChain.layer]) : null;
        return (tokenAddress ?? zeroAddress) as Address;
      }),
    [filteredTokens, fromChain?.layer],
  );

  const { data } = useTokenPrices(tokenAddresses, fromChain?.id);

  const handleTokenClick = useCallback(
    (token: TokenInfo) => {
      setSelectedToken(token);
      onCloseModal();
    },
    [onCloseModal, setSelectedToken],
  );

  const getTokenPrice = useCallback(
    (token: TokenInfo): number | undefined => {
      if (fromChain && !fromChain.testnet && !isEmptyObject(data)) {
        const tokenAddress = (safeGetAddress(token[fromChain.layer]) || zeroAddress).toLowerCase();
        return data[tokenAddress];
      }
      return undefined;
    },
    [data, fromChain],
  );

  const normalizeInput = useCallback((input: string): string => {
    return isAddress(input) ? getAddress(input) : input.toLowerCase();
  }, []);

  const handleSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchQuery(normalizeInput(e.target.value));
    },
    [normalizeInput],
  );

  return (
    <Modal title="Select a token" isOpen={isModalOpen} onClose={onCloseModal} isDrawer={isMobile}>
      <div className={styles["modal-inner"]}>
        <div className={styles["input-wrapper"]}>
          <SearchIcon />
          <input
            type="text"
            placeholder="Search by token name, symbol or address"
            onChange={handleSearchChange}
            value={searchQuery}
          />
        </div>
        <div className={styles["list-token"]}>
          {filteredTokens.length > 0 ? (
            filteredTokens.map((token: TokenInfo, index: number) => {
              return (
                <TokenDetails
                  token={token}
                  onTokenClick={handleTokenClick}
                  key={`token-details-${index}`}
                  tokenPrice={getTokenPrice(token)}
                  currency={currency}
                />
              );
            })
          ) : (
            <div className={styles["not-found"]}>
              <p>Sorry, there are no results for that term. Please enter a valid token name or address.</p>
            </div>
          )}
        </div>
      </div>
    </Modal>
  );
}
