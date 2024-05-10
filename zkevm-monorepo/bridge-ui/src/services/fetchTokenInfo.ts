import { Address } from 'viem';
import { fetchToken } from '@wagmi/core';
import { sepolia, linea, mainnet, Chain } from 'viem/chains';
import log from 'loglevel';
import fetchERC20Image from '@/services/fetchERC20Image';
import { TokenInfo, TokenType } from '@/config/config';
import { NetworkType } from '@/contexts/chain.context';
import { lineaSepolia } from '@/utils/SepoliaChain';

const fetchTokenInfo = async (
  tokenAddress: Address,
  networkType: NetworkType,
  activeChain?: Chain,
): Promise<TokenInfo | undefined> => {
  let erc20;
  let chainFound;

  if (!chainFound) {
    const chains: Chain[] = networkType === NetworkType.SEPOLIA ? [lineaSepolia, sepolia] : [linea, mainnet];

    // Put the activeChain arg at the begining to take it as priority
    if (activeChain) chains.unshift(activeChain);

    for (const chain of chains) {
      try {
        erc20 = await fetchToken({
          address: tokenAddress,
          chainId: chain.id,
        });
        if (erc20.name) {
          // Found the token if no errors with fetchToken
          chainFound = chain;
          break;
        }
      } catch (err) {
        continue;
      }
    }
  }

  if (!erc20 || !chainFound || !erc20.name) {
    return;
  }

  const L1Token = chainFound.id === mainnet.id || chainFound.id === sepolia.id;

  // Fetch image
  const name = erc20.name;
  const image = await fetchERC20Image(name);

  try {
    return {
      name,
      symbol: erc20.symbol,
      decimals: erc20.decimals,
      L1: L1Token ? erc20.address : null,
      L2: !L1Token ? erc20.address : null,
      image,
      type: TokenType.ERC20,
      UNKNOWN: null,
      isDefault: false,
    };
  } catch (err) {
    log.error(err);
    return;
  }
};

export default fetchTokenInfo;
