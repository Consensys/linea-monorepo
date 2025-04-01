import { useEffect } from "react";
import { useConfigStore } from "./configStoreProvider";
import { watchAccount } from "@wagmi/core";
import { useChainStore } from "./chainStoreProvider";
import { config } from "@/lib/wagmi";
import { Chain, ChainLayer } from "@/types";

export function ChainUpdater() {
  const showTestnet = useConfigStore((state) => state.showTestnet);
  const chains = useChainStore((state) => state.chains);
  const setFromChain = useChainStore((state) => state.setFromChain);
  const setToChain = useChainStore((state) => state.setToChain);
  const setAvailableChains = useChainStore((state) => state.setAvailableChains);

  useEffect(() => {
    setAvailableChains(showTestnet);

    const l1Chain = chains.find((c) => c.testnet === showTestnet && c.layer === ChainLayer.L1);
    const l2Chain = chains.find((c) => c.testnet === showTestnet && c.layer === ChainLayer.L2);
    setFromChain(l1Chain);
    setToChain(l2Chain);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [showTestnet]);

  useEffect(() => {
    const unwatch = watchAccount(config, {
      onChange(account) {
        const chain = chains.find((chain: Chain) => chain.id === account?.chain?.id);

        if (!chain) {
          return;
        }

        setFromChain(chain);

        if (chain.testnet) {
          setToChain(chains?.find((c: Chain) => c.testnet && c.layer !== chain.layer));
        } else {
          setToChain(chains?.find((c: Chain) => !c.testnet && c.layer !== chain.layer));
        }
      },
    });

    return () => {
      unwatch();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  return null;
}
