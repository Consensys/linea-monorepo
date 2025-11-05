import { useAccount, useConfig } from "wagmi";
import { useCallback, useMemo } from "react";
import { watchAccount, getAccount } from "@wagmi/core";
import { resolveWalletConnectorIcon, NetworkWithTokens, NetworkType } from "@layerswap/widget";
import {
  WalletConnectionProvider,
  Wallet,
  WalletConnectionProviderProps,
  InternalConnector,
} from "@layerswap/widget/types";
import { useWeb3AuthConnect, useWeb3AuthDisconnect } from "@web3auth/modal/react";

export default function useEVM({ networks }: WalletConnectionProviderProps): WalletConnectionProvider {
  const name = "EVM";
  const id = "evm";

  // wagmi
  const wagmiConfig = useConfig();
  const { connector: activeConnector, address: activeAddress, isConnected } = useAccount();
  const { connect } = useWeb3AuthConnect();
  const { disconnect } = useWeb3AuthDisconnect();

  // Gather the EVM‐type network names
  const evmNetworkNames = useMemo(
    () => networks.filter((n) => n.type === NetworkType.EVM).map((n) => n.name),
    [networks],
  );

  // Supported-networks
  const supportedNetworks = useMemo(
    () => ({
      asSource: evmNetworkNames,
      autofill: evmNetworkNames,
      withdrawal: evmNetworkNames,
    }),
    [evmNetworkNames],
  );

  // connectWallet: trigger Web3Auth connection
  const connectWallet = useCallback(async (): Promise<Wallet | undefined> => {
    try {
      // Disconnect if already connected
      if (isConnected) {
        await disconnect();
      }

      // Connect
      await connect();

      // Wait for account address to become available
      const address = await waitForAccountAddress(wagmiConfig);

      if (!address) {
        return undefined;
      }

      return resolveWallet({
        connectorId: "web3auth",
        connectorName: "Web3Auth",
        address,
        isActive: true,
        networks,
        supportedNetworks,
        disconnect: async () => disconnect(),
        providerName: name,
      });
    } catch (error) {
      console.error("Failed to connect wallet:", error);
      return undefined;
    }
  }, [isConnected, connect, wagmiConfig, networks, supportedNetworks, disconnect]);

  // Logout
  const disconnectWallets = useCallback(async () => {
    disconnect();
  }, [disconnect]);

  // Map connected wallet to Layerswap Wallet shape
  const connectedWallets: Wallet[] = useMemo(() => {
    if (!isConnected || !activeAddress || !activeConnector) {
      return [];
    }

    const wallet = resolveWallet({
      connectorId: activeConnector.id,
      connectorName: activeConnector.name,
      address: activeAddress,
      isActive: true,
      networks,
      supportedNetworks,
      disconnect: disconnectWallets,
      providerName: name,
    });

    return wallet ? [wallet] : [];
  }, [isConnected, activeAddress, activeConnector, networks, supportedNetworks, disconnectWallets]);

  const logo = networks.find((n) => n.name.toLowerCase().includes("linea"))?.logo;
  const availableConnectors: InternalConnector[] = [
    {
      id: "web3auth",
      name: "Web3Auth",
      icon: logo,
      providerName: name,
    },
  ];

  return {
    connectWallet,
    availableWalletsForConnect: availableConnectors,
    activeWallet: connectedWallets.find((w) => w.isActive),
    connectedWallets,
    asSourceSupportedNetworks: supportedNetworks.asSource,
    autofillSupportedNetworks: supportedNetworks.autofill,
    withdrawalSupportedNetworks: supportedNetworks.withdrawal,
    name,
    id,
    providerIcon: logo,
  };
}

/** Reusable helper to turn a wagmi connector + address into Layerswap `Wallet` shape */
function resolveWallet(props: {
  connectorId: string;
  connectorName: string;
  address: string;
  isActive: boolean;
  networks: NetworkWithTokens[];
  supportedNetworks: {
    asSource: string[];
    autofill: string[];
    withdrawal: string[];
  };
  disconnect: () => Promise<void>;
  providerName: string;
}): Wallet | undefined {
  const { connectorId, connectorName, address, isActive, networks, supportedNetworks, disconnect, providerName } =
    props;

  if (!address) return;

  const displayName = `${connectorName} – ${providerName}`;
  const networkIcon = networks.find((n) => n.name.toLowerCase().includes("linea"))?.logo;

  return {
    id: connectorId,
    isActive,
    address,
    addresses: [address],
    displayName,
    providerName,
    icon: resolveWalletConnectorIcon({ connector: connectorName, address }),
    disconnect: () => disconnect(),
    asSourceSupportedNetworks: supportedNetworks.asSource,
    autofillSupportedNetworks: supportedNetworks.autofill,
    withdrawalSupportedNetworks: supportedNetworks.withdrawal,
    networkIcon,
  };
}

/**
 * Helper function to wait for account address to become available
 * @param wagmiConfig - Wagmi configuration
 * @param timeout - Maximum time to wait in milliseconds (default: 30000)
 * @returns Promise that resolves with the account address or rejects on timeout
 */
function waitForAccountAddress(wagmiConfig: ReturnType<typeof useConfig>, timeout: number = 30000): Promise<string> {
  return new Promise((resolve, reject) => {
    // Check immediately in case account is already available
    const account = getAccount(wagmiConfig);
    if (account.address) {
      resolve(account.address);
      return;
    }

    // Watch for account changes
    const unwatch = watchAccount(wagmiConfig, {
      onChange(account) {
        if (account.address) {
          unwatch();
          clearTimeout(timeoutId);
          resolve(account.address);
        }
      },
    });

    // Set up timeout
    const timeoutId = setTimeout(() => {
      unwatch();
      reject(new Error("Timeout waiting for wallet connection"));
    }, timeout);
  });
}
