'use client';

import React, { useContext } from 'react';
import Image from 'next/image';
import { useFormContext } from 'react-hook-form';
import { useAccount, useBalance } from 'wagmi';
import classNames from 'classnames';

import { config } from '@/config';
import { formatBalance } from '@/utils/format';
import { TokenInfo, TokenType } from '@/config/config';
import { ChainContext, NetworkLayer, NetworkType } from '@/contexts/chain.context';

interface TokenDetailsProps {
  token: TokenInfo;
  onTokenClick: (token: TokenInfo) => void;
}
export default function TokenDetails({ token, onTokenClick }: TokenDetailsProps) {
  const { address } = useAccount();
  const context = useContext(ChainContext);
  const { networkLayer, fromChain, setToken, setTokenBridgeAddress, networkType } = context;

  const tokenNotFromCurrentLayer = !token[networkLayer] && token.type !== TokenType.ETH;

  // Form
  const { setValue, clearErrors } = useFormContext();

  const { data: balance } = useBalance({
    address,
    token: token[networkLayer] ?? undefined,
    chainId: fromChain?.id,
    watch: true,
  });

  return (
    <button
      id={`token-details-${token.symbol}-btn`}
      className={classNames(
        'flex items-center justify-between w-full gap-5 px-8 py-3 mt-3 bg-transparent border-0 hover:bg-slate-900/20',
        {
          'btn-disabled': tokenNotFromCurrentLayer,
        },
      )}
      disabled={tokenNotFromCurrentLayer}
      onClick={() => {
        if (networkLayer !== NetworkLayer.UNKNOWN && token && networkType !== NetworkType.WRONG_NETWORK) {
          setValue('amount', '');
          clearErrors('amount');
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
        <Image src={token.image} alt={token.name} width={40} height={40} className="rounded-full" />
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
