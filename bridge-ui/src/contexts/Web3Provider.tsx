"use client";

import { ReactNode, useEffect } from "react";
import { useWalletDetection } from "./WalletDetectionProvider";
import { useWeb3Auth, Web3AuthContextConfig, Web3AuthProvider } from "@web3auth/modal/react";
import { WagmiProvider } from "@web3auth/modal/react/wagmi";
import { WEB3AUTH_NETWORK } from "@web3auth/modal";
import { ADAPTER_EVENTS } from "@web3auth/base";
import useGTM from "@/hooks/useGtm";
import { useCachedIdentityToken } from "@/hooks/useCachedIdentityToken";
import { isProd } from "../../next.config.mjs";
import { config as appConfig } from "@/config";

interface DynamicProviderProps {
  children: ReactNode;
}

const clientId = appConfig.web3AuthClientId;

const web3AuthContextConfig: Web3AuthContextConfig = {
  web3AuthOptions: {
    clientId,
    web3AuthNetwork: isProd ? WEB3AUTH_NETWORK.SAPPHIRE_MAINNET : WEB3AUTH_NETWORK.SAPPHIRE_DEVNET,
    defaultChainId: appConfig.e2eTestMode ? "0x1E2EAAC" : "0xe708", // L2 local chain or Linea Mainnet
    uiConfig: {
      appUrl: "https://linea.build/hub/bridge",
      displayErrorsOnModal: true,
    },
    modalConfig: {
      connectors: {
        auth: {
          showOnModal: false,
          label: "Web3Auth",
        },
      },
    },
  },
};

function Web3AuthEventBridge() {
  const { web3Auth } = useWeb3Auth();
  const { walletsInstalled } = useWalletDetection();
  const { trackEvent } = useGTM();
  const { clearTokenCache } = useCachedIdentityToken();

  useEffect(() => {
    if (walletsInstalled?.length > 0) {
      trackEvent({
        event: "wallets_installed",
        wallets: walletsInstalled?.join(","),
      });
    }
  }, [trackEvent, walletsInstalled]);

  useEffect(() => {
    if (!web3Auth || !trackEvent) return;

    const onConnecting = () => {
      trackEvent({
        event: "before_all_clicks",
        click_text_before: "Connect wallet",
        gtm_human_readable_id_before: "gtm-connect-wallet-button",
      });
    };

    const onConnected = (args: { connector: string }) => {
      trackEvent({
        event: "wallet_connected",
        wallet_connected: args?.connector,
        wallets_installed: walletsInstalled?.join(","),
      });
    };

    const onDisconnected = () => {
      clearTokenCache();
      trackEvent({
        event: "before_all_clicks",
        click_text_before: "Connect wallet",
        gtm_human_readable_id_before: "gtm-connect-wallet-button",
      });
    };

    const onErrored = () => {
      clearTokenCache();
      trackEvent({
        event: "before_all_clicks",
        click_text_before: "Connect wallet",
        gtm_human_readable_id_before: "gtm-connect-wallet-button",
      });
    };

    web3Auth.on(ADAPTER_EVENTS.CONNECTING, onConnecting);
    web3Auth.on(ADAPTER_EVENTS.CONNECTED, onConnected);
    web3Auth.on(ADAPTER_EVENTS.DISCONNECTED, onDisconnected);
    web3Auth.on(ADAPTER_EVENTS.ERRORED, onErrored);

    return () => {
      web3Auth.off(ADAPTER_EVENTS.CONNECTING, onConnecting);
      web3Auth.off(ADAPTER_EVENTS.CONNECTED, onConnected);
      web3Auth.off(ADAPTER_EVENTS.DISCONNECTED, onDisconnected);
      web3Auth.off(ADAPTER_EVENTS.ERRORED, onErrored);
    };
  }, [web3Auth, trackEvent, walletsInstalled, clearTokenCache]);

  return null;
}

export function Web3Provider({ children }: DynamicProviderProps) {
  return (
    <Web3AuthProvider config={web3AuthContextConfig}>
      <Web3AuthEventBridge />
      <WagmiProvider>{children}</WagmiProvider>
    </Web3AuthProvider>
  );
}
