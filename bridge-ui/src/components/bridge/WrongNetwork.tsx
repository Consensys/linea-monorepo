import Image from "next/image";
import { linea, mainnet } from "viem/chains";
import { wagmiConfig } from "@/config";
import { switchChain } from "@wagmi/core";

export default function WrongNetwork() {
  const switchNetwork = async (id: number) => {
    await switchChain(wagmiConfig, {
      chainId: id,
    });
  };

  return (
    <div className="card-body">
      <div className="space-y-5">
        <div>Wrong network</div>
        <ul className="space-y-5">
          <li>
            <button className="btn" onClick={() => switchNetwork(mainnet.id)}>
              <Image
                src={"/images/logo/ethereum.svg"}
                alt="Ethereum"
                width={0}
                height={0}
                style={{ width: "12px", height: "auto" }}
              />{" "}
              Mainnet
            </button>
          </li>
          <li>
            <button className="btn" onClick={() => switchNetwork(linea.id)}>
              <Image
                src={"/images/logo/linea-mainnet.svg"}
                alt="Linea"
                width={0}
                height={0}
                style={{ width: "18px", height: "auto" }}
              />{" "}
              Linea
            </button>
          </li>
        </ul>
      </div>
    </div>
  );
}
