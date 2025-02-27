import { Web3Provider } from "@/contexts/web3.context";
import { ModalProvider } from "@/contexts/modal.context";
import { TokenStoreProvider } from "@/stores/tokenStoreProvider";
import { getTokenConfig } from "@/services/tokenService";

type ProvidersProps = {
  children: JSX.Element;
};

async function getTokenStoreInitialState() {
  const tokensList = await getTokenConfig();

  return { tokensList, selectedToken: tokensList.MAINNET[0] };
}

export async function Providers({ children }: ProvidersProps) {
  const tokensStoreInitialState = await getTokenStoreInitialState();

  return (
    <Web3Provider>
      <TokenStoreProvider initialState={tokensStoreInitialState}>
        <ModalProvider>{children}</ModalProvider>
      </TokenStoreProvider>
    </Web3Provider>
  );
}
