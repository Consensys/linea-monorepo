"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { http, type Chain } from "viem";
import { WagmiProvider, createConfig, useAccount, useConnect, useDisconnect } from "wagmi";
import { injected } from "wagmi/connectors";

import { LineaWordmark } from "./components/LineaWordmark";

declare global {
  interface Window {
    ethereum?: {
      on?: (event: string, handler: (...args: unknown[]) => void) => void;
      removeListener?: (event: string, handler: (...args: unknown[]) => void) => void;
      request: <T>(request: { method: string; params?: unknown[] }) => Promise<T>;
    };
  }
}

type ChainMetadata = {
  chainId: number;
  chainName: string;
  rpcUrls: string[];
  blockExplorerUrls: string[];
  nativeCurrency: {
    name: string;
    symbol: string;
    decimals: number;
  };
};

type TransactionPrompt = {
  id: string;
  label: string;
  description: string;
  createdAt: string;
  request: {
    to?: string;
    data?: string;
    value?: string;
    gas?: string;
    gasPrice?: string;
    maxFeePerGas?: string;
    maxPriorityFeePerGas?: string;
    nonce?: string;
    type?: string;
    chainId?: string;
  };
  deploymentDetails?: {
    contractName?: string;
    constructorArgs?: unknown;
    initializerArgs?: unknown;
    proxyOptions?: string;
    notes?: string;
  } | null;
};

type SessionState = {
  sessionId: string;
  deployFile: string;
  networkName: string;
  chain: ChainMetadata;
  wallet: {
    address: string;
    chainId: number;
    connectedAt: string;
  } | null;
  pendingRequest: TransactionPrompt | null;
  startedAt: string;
};

const knownChains: readonly [Chain, ...Chain[]] = [
  {
    id: 1,
    name: "Ethereum Mainnet",
    nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
    rpcUrls: { default: { http: ["https://mainnet.infura.io/v3"] } },
    blockExplorers: { default: { name: "Etherscan", url: "https://etherscan.io" } },
  },
  {
    id: 11155111,
    name: "Sepolia",
    nativeCurrency: { name: "Sepolia Ether", symbol: "SEP", decimals: 18 },
    rpcUrls: { default: { http: ["https://sepolia.infura.io/v3"] } },
    blockExplorers: { default: { name: "Etherscan", url: "https://sepolia.etherscan.io" } },
  },
  {
    id: 560048,
    name: "Hoodi",
    nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
    rpcUrls: { default: { http: ["https://hoodi.infura.io/v3"] } },
    blockExplorers: { default: { name: "Etherscan", url: "https://hoodi.etherscan.io" } },
  },
  {
    id: 31648428,
    name: "Linea Local L1 (Docker)",
    nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
    rpcUrls: { default: { http: ["http://127.0.0.1:8445"] } },
    blockExplorers: { default: { name: "Explorer", url: "" } },
  },
  {
    id: 59139,
    name: "Linea Devnet (hosted)",
    nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
    rpcUrls: { default: { http: ["https://rpc.devnet.linea.build"] } },
    blockExplorers: { default: { name: "Explorer", url: "" } },
  },
  {
    id: 59141,
    name: "Linea Sepolia",
    nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
    rpcUrls: { default: { http: ["https://linea-sepolia.infura.io/v3"] } },
    blockExplorers: { default: { name: "LineaScan", url: "https://sepolia.lineascan.build" } },
  },
  {
    id: 59144,
    name: "Linea Mainnet",
    nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
    rpcUrls: { default: { http: ["https://linea-mainnet.infura.io/v3"] } },
    blockExplorers: { default: { name: "LineaScan", url: "https://lineascan.build" } },
  },
];

const queryClient = new QueryClient();

const wagmiConfig = createConfig({
  chains: knownChains,
  connectors: [injected()],
  transports: Object.fromEntries(knownChains.map((chain) => [chain.id, http()])) as Record<
    number,
    ReturnType<typeof http>
  >,
});

/**
 * Must match `DEPLOY_UI_SESSION_TOKEN_HEADER` in `contracts/scripts/hardhat/deployment-ui.ts`
 * (HTTP headers are case-insensitive; Node normalizes to lowercase).
 */
const DEPLOY_UI_SESSION_TOKEN_HEADER = "X-Deploy-Ui-Session-Token";

async function postJson(url: string, payload: unknown, sessionSecret: string) {
  const response = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      [DEPLOY_UI_SESSION_TOKEN_HEADER]: sessionSecret,
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request to ${url} failed with ${response.status}`);
  }
}

function ContractsDeployUiPage() {
  const searchParams = useSearchParams();
  const apiBaseUrl = searchParams.get("apiBaseUrl");
  const sessionSecretFromUrl =
    searchParams.get("sessionToken") ?? searchParams.get("bridgeToken");
  const [sessionSecret, setSessionSecret] = useState<string | null>(null);
  const [sessionAuthReady, setSessionAuthReady] = useState(false);
  const { address, isConnected } = useAccount();
  const { connect, connectors, isPending: isConnecting, error: connectError } = useConnect();
  const { disconnect } = useDisconnect();
  const [session, setSession] = useState<SessionState | null>(null);
  const [statusMessage, setStatusMessage] = useState<string>("Waiting for deployment session...");
  const [actionError, setActionError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  /** Wagmi reconnect differs SSR vs first client paint — gate wallet UI until mount. */
  const [walletUiReady, setWalletUiReady] = useState(false);
  const [sessionEnded, setSessionEnded] = useState(false);
  const hadSuccessfulBridgeFetch = useRef(false);

  const injectedConnector = connectors[0];

  useEffect(() => {
    setWalletUiReady(true);
  }, []);

  useEffect(() => {
    if (!apiBaseUrl) {
      setSessionSecret(null);
      setSessionAuthReady(true);
      return;
    }

    const storageKey = `deployUiSessionSecret:${apiBaseUrl}`;
    const legacyStorageKey = `lineaDeployUiBridgeToken:${apiBaseUrl}`;

    if (sessionSecretFromUrl) {
      sessionStorage.setItem(storageKey, sessionSecretFromUrl);
      setSessionSecret(sessionSecretFromUrl);
      const nextUrl = new URL(window.location.href);
      nextUrl.searchParams.delete("sessionToken");
      nextUrl.searchParams.delete("bridgeToken");
      window.history.replaceState({}, "", `${nextUrl.pathname}${nextUrl.search}${nextUrl.hash}`);
      setSessionAuthReady(true);
      return;
    }

    setSessionSecret(sessionStorage.getItem(storageKey) ?? sessionStorage.getItem(legacyStorageKey));
    setSessionAuthReady(true);
  }, [apiBaseUrl, sessionSecretFromUrl]);

  const sessionUrl = useMemo(() => {
    if (!apiBaseUrl) {
      return undefined;
    }

    return `${apiBaseUrl}/api/session`;
  }, [apiBaseUrl]);

  useEffect(() => {
    if (!sessionAuthReady) {
      return;
    }

    if (!apiBaseUrl) {
      setActionError("Missing apiBaseUrl query parameter.");
      return;
    }

    if (!sessionSecret) {
      setActionError(
        "Missing session token. Open this UI using the full URL from Hardhat (DEPLOY_WITH_UI), not a bookmark without sessionToken.",
      );
      return;
    }

    if (!sessionUrl) {
      return;
    }

    if (sessionEnded) {
      return;
    }

    let cancelled = false;
    const load = async () => {
      try {
        const response = await fetch(sessionUrl, {
          headers: { [DEPLOY_UI_SESSION_TOKEN_HEADER]: sessionSecret },
        });
        if (!response.ok) {
          if (!cancelled && hadSuccessfulBridgeFetch.current) {
            setSessionEnded(true);
            setStatusMessage("Deploy session ended (bridge closed).");
          } else if (!cancelled && !hadSuccessfulBridgeFetch.current) {
            setActionError(`Session request failed (${response.status}).`);
          }
          return;
        }
        const data = (await response.json()) as SessionState;
        if (!cancelled) {
          hadSuccessfulBridgeFetch.current = true;
          setSession(data);
          setStatusMessage(`Connected to ${data.deployFile} on ${data.networkName}`);
          setActionError(null);
        }
      } catch {
        if (!cancelled) {
          if (hadSuccessfulBridgeFetch.current) {
            setSessionEnded(true);
            setStatusMessage("Deploy session ended (bridge closed).");
          } else {
            setActionError("Failed to reach the deploy bridge. Check apiBaseUrl and that Hardhat started the session.");
          }
        }
      }
    };

    void load();
    const interval = window.setInterval(() => {
      void load();
    }, 1500);

    return () => {
      cancelled = true;
      window.clearInterval(interval);
    };
  }, [sessionUrl, sessionEnded, sessionAuthReady, sessionSecret, apiBaseUrl]);

  useEffect(() => {
    if (!apiBaseUrl || !sessionSecret || !isConnected || !address || !window.ethereum || sessionEnded) {
      return;
    }

    const syncWalletState = async () => {
      try {
        const currentChainIdHex = await window.ethereum?.request<string>({ method: "eth_chainId" });
        if (!currentChainIdHex) {
          return;
        }

        await postJson(
          `${apiBaseUrl}/api/wallet`,
          {
            address,
            chainId: parseInt(currentChainIdHex, 16),
          },
          sessionSecret,
        );
        setActionError(null);
      } catch (error) {
        setActionError((error as Error).message ?? "Failed to register wallet with the deploy bridge.");
      }
    };

    void syncWalletState();

    const accountsChanged = () => {
      void syncWalletState();
    };
    const chainChanged = () => {
      void syncWalletState();
    };

    window.ethereum.on?.("accountsChanged", accountsChanged);
    window.ethereum.on?.("chainChanged", chainChanged);

    return () => {
      window.ethereum?.removeListener?.("accountsChanged", accountsChanged);
      window.ethereum?.removeListener?.("chainChanged", chainChanged);
    };
  }, [address, apiBaseUrl, sessionSecret, isConnected, sessionEnded]);

  const ensureTargetChain = async (chain: ChainMetadata) => {
    if (!window.ethereum) {
      throw new Error("No injected wallet was found in this browser.");
    }

    const targetChainIdHex = `0x${chain.chainId.toString(16)}`;
    const currentChainId = await window.ethereum.request<string>({ method: "eth_chainId" });

    if (currentChainId === targetChainIdHex) {
      return;
    }

    try {
      await window.ethereum.request({
        method: "wallet_switchEthereumChain",
        params: [{ chainId: targetChainIdHex }],
      });
      return;
    } catch (error) {
      const switchError = error as { code?: number; message?: string };
      if (switchError.code !== 4902) {
        throw new Error(switchError.message ?? "Failed to switch to the deployment chain.");
      }
    }

    await window.ethereum.request({
      method: "wallet_addEthereumChain",
      params: [
        {
          chainId: targetChainIdHex,
          chainName: chain.chainName,
          rpcUrls: chain.rpcUrls,
          blockExplorerUrls: chain.blockExplorerUrls,
          nativeCurrency: chain.nativeCurrency,
        },
      ],
    });
    await window.ethereum.request({
      method: "wallet_switchEthereumChain",
      params: [{ chainId: targetChainIdHex }],
    });
  };

  const approvePendingRequest = async () => {
    if (!apiBaseUrl || !sessionSecret || !session?.pendingRequest || !window.ethereum || !address) {
      return;
    }

    setIsSubmitting(true);
    setActionError(null);

    try {
      await ensureTargetChain(session.chain);

      const hash = await window.ethereum.request<string>({
        method: "eth_sendTransaction",
        params: [
          {
            ...session.pendingRequest.request,
            from: address,
          },
        ],
      });

      await postJson(
        `${apiBaseUrl}/api/respond`,
        {
          requestId: session.pendingRequest.id,
          hash,
          from: address,
          chainId: session.chain.chainId,
        },
        sessionSecret,
      );
      setStatusMessage(`Submitted ${session.pendingRequest.label}`);
    } catch (error) {
      const message = (error as Error).message;
      setActionError(message);

      if (session.pendingRequest) {
        await postJson(
          `${apiBaseUrl}/api/error`,
          {
            requestId: session.pendingRequest.id,
            message,
          },
          sessionSecret,
        );
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="deploy-shell">
      <header className="deploy-header">
        <div className="deploy-header__brand">
          <LineaWordmark className="deploy-header__logo" />
          <span className="deploy-header__badge">Contracts · local</span>
        </div>
      </header>

      <main className="deploy-main">
        <section className="deploy-card deploy-card--hero">
          <h1 className="deploy-card__title">Deploy with your wallet</h1>
          <p className="deploy-card__subtitle">{statusMessage}</p>
        </section>

        {sessionEnded ? (
          <section className="deploy-card deploy-card--muted">
            <h2>Session disconnected</h2>
            <p>
              The Hardhat deploy bridge has stopped. Polling is disabled. If you are not running another deployment,
              you can close this browser tab.
            </p>
          </section>
        ) : null}

        <section className="deploy-card">
          <h2>Wallet</h2>
          <p>
            {!walletUiReady
              ? "Preparing wallet connection…"
              : isConnected && address
                ? `Connected as ${address}`
                : "Connect the wallet that should sign this deploy file."}
          </p>
          <div className="deploy-actions">
            {!walletUiReady ? null : !isConnected && injectedConnector ? (
              <button
                type="button"
                className="deploy-btn deploy-btn--primary"
                disabled={isConnecting}
                onClick={() => connect({ connector: injectedConnector })}
              >
                {isConnecting ? "Connecting..." : "Connect wallet"}
              </button>
            ) : null}
            {!walletUiReady ? null : isConnected ? (
              <button type="button" className="deploy-btn deploy-btn--secondary" onClick={() => disconnect()}>
                Disconnect
              </button>
            ) : null}
          </div>
          {connectError ? <p className="deploy-connect-error">{connectError.message}</p> : null}
        </section>

        <section className="deploy-card">
          <h2>Deployment target</h2>
          <p className="deploy-meta-row">
            <strong>Deploy file:</strong> {session?.deployFile ?? "Loading…"}
          </p>
          <p className="deploy-meta-row">
            <strong>Network:</strong> {session?.networkName ?? "Loading…"}
          </p>
          <p className="deploy-meta-row">
            <strong>Chain:</strong> {session?.chain.chainName ?? "Loading…"}
          </p>
          <p className="deploy-meta-row">
            <strong>Chain ID:</strong> {session?.chain.chainId ?? "Loading…"}
          </p>
        </section>

        <section className="deploy-card">
          <h2>Pending request</h2>
          {session?.pendingRequest ? (
            <>
              <p className="deploy-pending-label">{session.pendingRequest.label}</p>
              <p>{session.pendingRequest.description}</p>
              {session.pendingRequest.deploymentDetails ? (
                <>
                  <h3>Deployment parameters (Hardhat)</h3>
                  <pre className="deploy-code">
                    {JSON.stringify(session.pendingRequest.deploymentDetails, null, 2)}
                  </pre>
                </>
              ) : null}
              <h3>Raw transaction request</h3>
              <pre className="deploy-code">{JSON.stringify(session.pendingRequest.request, null, 2)}</pre>
              <div className="deploy-actions">
                <button
                  type="button"
                  className="deploy-btn deploy-btn--primary"
                  disabled={isSubmitting || !walletUiReady || !isConnected}
                  onClick={() => void approvePendingRequest()}
                >
                  {isSubmitting ? "Submitting…" : "Switch chain, sign, and send"}
                </button>
              </div>
            </>
          ) : (
            <p>No transaction is waiting for approval yet.</p>
          )}
        </section>

        {actionError ? (
          <section className="deploy-card deploy-card--error">
            <h2>Error</h2>
            <p>{actionError}</p>
          </section>
        ) : null}
      </main>
    </div>
  );
}

export default function Page() {
  return (
    <WagmiProvider config={wagmiConfig}>
      <QueryClientProvider client={queryClient}>
        <ContractsDeployUiPage />
      </QueryClientProvider>
    </WagmiProvider>
  );
}
