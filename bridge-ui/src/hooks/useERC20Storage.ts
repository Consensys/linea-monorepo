import { useCallback } from "react";
import { NetworkType, TokenInfo, TokenType } from "@/config/config";
import { safeGetAddress } from "@/utils/format";
import { useTokenStore } from "@/stores/tokenStore";

const useERC20Storage = () => {
  const { storedTokens, setStoredTokens } = useTokenStore((state) => ({
    storedTokens: state.usersTokens,
    setStoredTokens: state.setUsersTokens,
  }));

  const getStoredToken = useCallback(
    (token: TokenInfo, networkType: NetworkType) => {
      if (networkType !== NetworkType.WRONG_NETWORK) {
        for (let i = 0; i < storedTokens[networkType].length; i++) {
          const storedToken = storedTokens[networkType][i];
          const l1Address = safeGetAddress(token.L1);
          const l2Address = safeGetAddress(token.L2);
          const storedL1Address = safeGetAddress(storedToken.L1);
          const storedL2Address = safeGetAddress(storedToken.L2);
          if (
            (l1Address && storedL1Address === l1Address) ||
            (l2Address && storedL2Address === l2Address) ||
            token.type === TokenType.ETH
          ) {
            return { storedToken, index: i };
          }
        }
      }
    },
    [storedTokens],
  );

  const updateOrInsertUserTokenList = useCallback(
    (token: TokenInfo, networkType: NetworkType) => {
      if (networkType !== NetworkType.WRONG_NETWORK && !token.isDefault) {
        const found = getStoredToken(token, networkType);
        if (found) {
          if (found.storedToken !== token) {
            // Update the token found
            storedTokens[networkType][found.index] = token;
            setStoredTokens(storedTokens);
          }
        } else {
          // Insert it to the list
          storedTokens[networkType].push(token);
          setStoredTokens(storedTokens);
        }
      }
    },
    [setStoredTokens, storedTokens, getStoredToken],
  );

  return { updateOrInsertUserTokenList };
};

export default useERC20Storage;
