import { http, createConfig } from "wagmi";
import { linea, lineaSepolia, mainnet, sepolia } from "wagmi/chains";

export const config = createConfig({
  chains: [mainnet, linea, lineaSepolia, sepolia],
  multiInjectedProviderDiscovery: false,
  transports: {
    [mainnet.id]: http(),
    [sepolia.id]: http(),
    [linea.id]: http(),
    [lineaSepolia.id]: http(),
  },
});

declare module "wagmi" {
  interface Register {
    config: typeof config;
  }
}
