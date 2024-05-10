import Image from 'next/image';
import { linea } from 'viem/chains';
import EthereumLogo from 'public/images/logo/ethereum.svg';
import LineaSepoliaLogo from 'public/images/logo/linea-sepolia.svg';
import LineaMainnetLogo from 'public/images/logo/linea-mainnet.svg';
import { lineaSepolia } from '@/utils/SepoliaChain';

type ChainLogoProps = {
  chainId: number;
};

const ChainLogo = ({ chainId }: ChainLogoProps) => {
  if (chainId === lineaSepolia.id) {
    return (
      <span>
        <Image src={LineaSepoliaLogo} alt="Linea Sepolia" width={18} height={18} />
      </span>
    );
  }

  if (chainId === linea.id) {
    return (
      <span>
        <Image src={LineaMainnetLogo} alt="Linea" width={18} height={18} />
      </span>
    );
  }

  return (
    <span>
      <Image src={EthereumLogo} alt="Ethereum" width={12} height={12} />
    </span>
  );
};

export default ChainLogo;
