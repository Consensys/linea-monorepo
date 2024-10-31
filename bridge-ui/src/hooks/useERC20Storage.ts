import { useCallback } from "react";
import { NetworkType, TokenInfo, TokenType } from "@/config/config";
import { safeGetAddress } from "@/utils/format";
import { useTokenStore } from "@/stores/tokenStoreProvider";
import { shallow } from "zustand/vanilla/shallow";

const useERC20Storage = () => {
  const tokensList = useTokenStore((state) => state.tokensList);
  const upsertToken = useTokenStore((state) => state.upsertToken);

  const getStoredToken = useCallback(
    (token: TokenInfo, networkType: NetworkType) => {
      if (networkType === NetworkType.WRONG_NETWORK) {
        return undefined;
      }

      const currentNetworkTokens = tokensList[networkType] || [];

      const index = currentNetworkTokens.findIndex((storedToken) => {
        const l1Address = safeGetAddress(token.L1);
        const l2Address = safeGetAddress(token.L2);
        const storedL1Address = safeGetAddress(storedToken.L1);
        const storedL2Address = safeGetAddress(storedToken.L2);

        return (
          (l1Address && storedL1Address === l1Address) ||
          (l2Address && storedL2Address === l2Address) ||
          token.type === TokenType.ETH
        );
      });

      if (index === -1) {
        return undefined;
      }

      return { storedToken: currentNetworkTokens[index], index };
    },
    [tokensList],
  );

  const updateOrInsertUserTokenList = useCallback(
    (token: TokenInfo, networkType: NetworkType) => {
      if (networkType === NetworkType.WRONG_NETWORK || token.isDefault) {
        return;
      }

      const found = getStoredToken(token, networkType);

      if (found) {
        if (!shallow(found.storedToken, token)) {
          upsertToken(token, networkType);
        } else {
          // No update needed if the token is the same
          return;
        }
      } else {
        upsertToken(token, networkType);
      }
    },
    [upsertToken, getStoredToken],
  );

  return { updateOrInsertUserTokenList };
};

export default useERC20Storage;
