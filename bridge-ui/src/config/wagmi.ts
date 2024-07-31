import { defaultWagmiConfig } from "@web3modal/wagmi/react/config";
import { cookieStorage, createStorage } from "wagmi";
import { http, injected } from "@wagmi/core";
import { mainnet, sepolia, linea, lineaSepolia } from "@wagmi/core/chains";
import { walletConnect, coinbaseWallet } from "@wagmi/connectors";
import { config } from "./config";

if (!config.walletConnectId) throw new Error("Project ID is not defined");

const metadata = {
  name: "Linea Bridge",
  description: `Linea Bridge is a bridge solution, providing secure and efficient cross-chain transactions between Layer 1 and Linea networks.
  Discover the future of blockchain interaction with Linea Bridge.`,
  url: "https://bridge.linea.build",
  icons: [],
};

const chains = [mainnet, sepolia, linea, lineaSepolia] as const;

export const wagmiConfig = defaultWagmiConfig({
  chains,
  projectId: config.walletConnectId,
  metadata,
  multiInjectedProviderDiscovery: true,
  ssr: true,
  enableEIP6963: true,
  connectors: [
    walletConnect({
      projectId: config.walletConnectId,
      showQrModal: false,
    }),
    injected({ shimDisconnect: true, target: "metaMask" }),
    coinbaseWallet({
      appName: "Linea Bridge",
    }),
  ],
  transports: {
    [mainnet.id]: http(`https://mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`),
    [sepolia.id]: http(`https://sepolia.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`),
    [linea.id]: http(`https://linea-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`),
    [lineaSepolia.id]: http(`https://linea-sepolia.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`),
  },
  storage: createStorage({
    storage: cookieStorage,
  }),
});
