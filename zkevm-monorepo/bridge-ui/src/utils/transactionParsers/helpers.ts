import { NetworkTokens, TokenInfo, TokenType } from '@/config/config';
import { safeGetAddress } from '../format';
import { NetworkType } from '@/contexts/chain.context';

export const findTokenByAddress = (
  tokenAddress: string,
  storedTokens: NetworkTokens,
  networkType: NetworkType,
): TokenInfo | undefined => {
  if (networkType === NetworkType.WRONG_NETWORK) return undefined;
  return storedTokens[networkType].find((token: TokenInfo) => {
    const l1Address = safeGetAddress(token.L1);
    const l2Address = safeGetAddress(token.L2);
    return l1Address === tokenAddress || l2Address === tokenAddress;
  });
};

export const findETHToken = (storedTokens: NetworkTokens, networkType: NetworkType): TokenInfo | undefined => {
  if (networkType === NetworkType.WRONG_NETWORK) return undefined;
  return storedTokens[networkType].find((token: TokenInfo) => token.type === TokenType.ETH);
};
