import { http, createConfig } from "wagmi";
import {
  linea,
  lineaSepolia,
  mainnet,
  sepolia,
  arbitrum,
  aurora,
  avalanche,
  base,
  blast,
  bsc,
  celo,
  cronos,
  fantom,
  gnosis,
  ink,
  mantle,
  mode,
  moonbeam,
  optimism,
  polygon,
  scroll,
  sei,
  sonic,
  zksync,
} from "wagmi/chains";

export const chains = [
  mainnet,
  linea,
  lineaSepolia,
  sepolia,
  arbitrum,
  aurora,
  avalanche,
  base,
  blast,
  bsc,
  celo,
  cronos,
  fantom,
  gnosis,
  ink,
  mantle,
  mode,
  moonbeam,
  optimism,
  polygon,
  scroll,
  sei,
  sonic,
  zksync,
] as const;
export const supportedChainIds = [mainnet.id, linea.id, lineaSepolia.id, sepolia.id] as const;

export type SupportedChainId = (typeof supportedChainIds)[number];

export const config = createConfig({
  chains,
  multiInjectedProviderDiscovery: false,
  transports: {
    [mainnet.id]: http(`https://mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [sepolia.id]: http(`https://sepolia.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [linea.id]: http(`https://linea-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [lineaSepolia.id]: http(`https://linea-sepolia.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [arbitrum.id]: http(`https://arbitrum-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [aurora.id]: http(`https://mainnet.aurora.dev`, { batch: true }),
    [avalanche.id]: http(`https://avalanche-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, {
      batch: true,
    }),
    [base.id]: http(`https://base-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [blast.id]: http(`https://blast-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [bsc.id]: http(`https://bsc-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [celo.id]: http(`https://celo-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [cronos.id]: http(`https://evm.cronos.org`, { batch: true }),
    [fantom.id]: http(`https://rpc.ankr.com/fantom`, { batch: true }),
    [gnosis.id]: http(`https://rpc.gnosischain.com`, { batch: true }),
    [ink.id]: http(`https://rpc-gel.inkonchain.com`, { batch: true }),
    [mantle.id]: http(`https://mantle-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [mode.id]: http(`https://mainnet.mode.network`, { batch: true }),
    [moonbeam.id]: http(`https://rpc.testnet.moonbeam.network`, { batch: true }),
    [optimism.id]: http(`https://optimism-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [polygon.id]: http(`https://polygon-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [scroll.id]: http(`https://scroll-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
    [sei.id]: http(`https://evm-rpc.sei-apis.com`, { batch: true }),
    [sonic.id]: http(`https://rpc.soniclabs.com`, { batch: true }),
    [zksync.id]: http(`https://zksync-mainnet.infura.io/v3/${process.env.NEXT_PUBLIC_INFURA_ID}`, { batch: true }),
  },
});

declare module "wagmi" {
  interface Register {
    config: typeof config;
  }
}
