import { Web3Provider } from "@/contexts/web3.context";
import { QueryProvider } from "@/contexts/query.context";

type ProvidersProps = {
  children: JSX.Element;
};

export async function Providers({ children }: ProvidersProps) {
  return (
    <QueryProvider>
      <Web3Provider>{children}</Web3Provider>
    </QueryProvider>
  );
}
