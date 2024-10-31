import { headers } from "next/headers";
import { cookieToInitialState } from "wagmi";
import { Web3Provider } from "@/contexts/web3.context";
import { ModalProvider } from "@/contexts/modal.context";
import { TokenStoreProvider } from "@/stores/tokenStoreProvider";
import { getTokenConfig } from "@/services/tokenService";
import { wagmiConfig } from "@/config";

type ProvidersProps = {
  children: JSX.Element;
};

export async function Providers({ children }: ProvidersProps) {
  const initialState = cookieToInitialState(wagmiConfig, headers().get("cookie"));
  const tokensList = await getTokenConfig();

  return (
    <TokenStoreProvider initialState={{ tokensList }}>
      <Web3Provider initialState={initialState}>
        <ModalProvider>{children}</ModalProvider>
      </Web3Provider>
    </TokenStoreProvider>
  );
}
