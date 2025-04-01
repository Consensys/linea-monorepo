import { ChainStoreProvider, ConfigStoreProvider, FormStoreProvider, TokenStoreProvider } from "@/stores";
import { getTokenConfig } from "@/services/tokenService";

async function getTokenStoreInitialState() {
  const tokensList = await getTokenConfig();

  return { tokensList, selectedToken: tokensList.MAINNET[0] };
}

export default async function Layout({ children }: { children: React.ReactNode }) {
  const tokensStoreInitialState = await getTokenStoreInitialState();

  return (
    <TokenStoreProvider initialState={tokensStoreInitialState}>
      <ConfigStoreProvider>
        <ChainStoreProvider>
          <FormStoreProvider initialToken={tokensStoreInitialState.tokensList.MAINNET[0]}>{children}</FormStoreProvider>
        </ChainStoreProvider>
      </ConfigStoreProvider>
    </TokenStoreProvider>
  );
}
