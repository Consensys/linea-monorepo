import { QueryProvider } from "@/contexts/query.context";
import { TokenStoreProvider } from "@/stores";
import { getTokenConfig } from "@/services/tokenService";
import { Web3Provider } from "@/contexts/Web3Provider";

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
    <QueryProvider>
      <Web3Provider>
        <TokenStoreProvider initialState={tokensStoreInitialState}>{children}</TokenStoreProvider>
      </Web3Provider>
    </QueryProvider>
  );
}
