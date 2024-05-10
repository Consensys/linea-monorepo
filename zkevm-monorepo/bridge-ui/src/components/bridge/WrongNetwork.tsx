import Image from 'next/image';
import { switchNetwork } from '@wagmi/core';
import EthereumLogo from 'public/images/logo/ethereum.svg';
import LineaMainnetLogo from 'public/images/logo/linea-mainnet.svg';
import { linea, mainnet } from 'viem/chains';

export default function WrongNetwork() {
  const switchChain = async (id: number) => {
    await switchNetwork({
      chainId: id,
    });
  };

  return (
    <div className="card-body">
      <div className="space-y-5">
        <div>Wrong network</div>
        <ul className="space-y-5">
          <li>
            <button className="btn" onClick={() => switchChain(mainnet.id)}>
              <Image src={EthereumLogo} alt="Ethereum" width={12} height={12} /> Mainnet
            </button>
          </li>
          <li>
            <button className="btn" onClick={() => switchChain(linea.id)}>
              <Image src={LineaMainnetLogo} alt="Linea" width={18} height={18} /> Linea
            </button>
          </li>
        </ul>
      </div>
    </div>
  );
}
