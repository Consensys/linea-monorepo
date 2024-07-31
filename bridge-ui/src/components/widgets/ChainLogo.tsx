import Image from "next/image";
import { linea, lineaSepolia } from "viem/chains";

type ChainLogoProps = {
  chainId: number;
};

const ChainLogo = ({ chainId }: ChainLogoProps) => {
  if (chainId === lineaSepolia.id) {
    return (
      <span>
        <Image
          src={"/images/logo/linea-sepolia.svg"}
          alt="Linea Sepolia"
          width={0}
          height={0}
          style={{ width: "18px", height: "auto" }}
        />
      </span>
    );
  }

  if (chainId === linea.id) {
    return (
      <span>
        <Image
          src={"/images/logo/linea-mainnet.svg"}
          alt="Linea"
          width={0}
          height={0}
          style={{ width: "18px", height: "auto" }}
        />
      </span>
    );
  }

  return (
    <span>
      <Image
        src={"/images/logo/ethereum.svg"}
        alt="Ethereum"
        width={0}
        height={0}
        style={{ width: "12px", height: "auto" }}
      />
    </span>
  );
};

export default ChainLogo;
