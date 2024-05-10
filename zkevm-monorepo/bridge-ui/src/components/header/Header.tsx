import { useContext } from 'react';
import Image from 'next/image';

import LineaLogo from 'public/images/logo/linea.svg';
import Wallet from './Wallet';
import Chains from './Chains';
import { useIsConnected } from '@/hooks';
import { ChainContext, NetworkType } from '@/contexts/chain.context';
import { UIContext } from '@/contexts/ui.context';

export default function Header() {
  // Hooks
  const isConnected = useIsConnected();

  // Context
  const context = useContext(ChainContext);
  const { networkType } = context;

  const { toggleShowBridge } = useContext(UIContext);

  return (
    <header className="navbar container py-4">
      <div className="flex-1">
        <button
          className="btn btn-ghost normal-case text-xl text-white w-32 md:w-52 -space-y-2 md:space-y-0"
          onClick={() => toggleShowBridge(false)}
        >
          <Image src={LineaLogo} alt="Linea" width={215} priority />
        </button>
        {networkType === NetworkType.SEPOLIA && (
          <div className="badge badge-primary badge-outline gap-2 ml-10">TESTNET</div>
        )}
      </div>
      <div className="flex-none">
        <ul className="menu menu-horizontal px-1">
          {isConnected && (
            <li>
              <Chains />
            </li>
          )}
          <li>
            <Wallet />
          </li>
        </ul>
      </div>
    </header>
  );
}
