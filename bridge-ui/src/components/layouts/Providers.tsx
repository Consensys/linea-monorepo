import { Web3Provider } from "@/contexts/web3.context";
import { ModalProvider } from "@/contexts/modal.context";
import { TokenStoreProvider } from "@/stores/tokenStoreProvider";
import { getTokenConfig } from "@/services/tokenService";

type ProvidersProps = {
  children: JSX.Element;
};

export async function Providers({ children }: ProvidersProps) {
  const tokensList = await getTokenConfig();

  return (
    <Web3Provider>
      <TokenStoreProvider initialState={{ tokensList }}>
        <ModalProvider>{children}</ModalProvider>
      </TokenStoreProvider>
    </Web3Provider>
  );
}
