"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { type CSSProperties, useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { http, type Chain } from "viem";
import { WagmiProvider, createConfig, useAccount, useConnect, useDisconnect } from "wagmi";
import { injected } from "wagmi/connectors";

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

function sectionStyle(): CSSProperties {
  return {
    border: "1px solid #243041",
    borderRadius: 12,
    backgroundColor: "#121a2f",
    padding: 20,
  };
}

function rawTransactionPreviewStyle(): CSSProperties {
  return {
    maxHeight: 320,
    overflow: "auto",
    overflowWrap: "anywhere",
    wordBreak: "break-word",
    whiteSpace: "pre-wrap",
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace',
    fontSize: 12,
    lineHeight: 1.45,
    borderRadius: 10,
    backgroundColor: "#0f172a",
    padding: 16,
    border: "1px solid #243041",
    marginBottom: 12,
  };
}

function buttonStyle(disabled = false): CSSProperties {
  return {
    border: 0,
    borderRadius: 10,
    padding: "10px 16px",
    fontWeight: 600,
    cursor: disabled ? "not-allowed" : "pointer",
    backgroundColor: disabled ? "#334155" : "#3b82f6",
    color: "#f8fafc",
  };
}

async function postJson(url: string, payload: unknown) {
  const response = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
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

  const sessionUrl = useMemo(() => {
    if (!apiBaseUrl) {
      return undefined;
    }

    return `${apiBaseUrl}/api/session`;
  }, [apiBaseUrl]);

  useEffect(() => {
    if (!sessionUrl) {
      setActionError("Missing apiBaseUrl query parameter.");
      return;
    }

    if (sessionEnded) {
      return;
    }

    let cancelled = false;
    const load = async () => {
      try {
        const response = await fetch(sessionUrl);
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
  }, [sessionUrl, sessionEnded]);

  useEffect(() => {
    if (!apiBaseUrl || !isConnected || !address || !window.ethereum || sessionEnded) {
      return;
    }

    const syncWalletState = async () => {
      const currentChainIdHex = await window.ethereum?.request<string>({ method: "eth_chainId" });
      if (!currentChainIdHex) {
        return;
      }

      await postJson(`${apiBaseUrl}/api/wallet`, {
        address,
        chainId: parseInt(currentChainIdHex, 16),
      });
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
  }, [address, apiBaseUrl, isConnected, sessionEnded]);

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
    if (!apiBaseUrl || !session?.pendingRequest || !window.ethereum || !address) {
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

      await postJson(`${apiBaseUrl}/api/respond`, {
        requestId: session.pendingRequest.id,
        hash,
        from: address,
        chainId: session.chain.chainId,
      });
      setStatusMessage(`Submitted ${session.pendingRequest.label}`);
    } catch (error) {
      const message = (error as Error).message;
      setActionError(message);

      if (session.pendingRequest) {
        await postJson(`${apiBaseUrl}/api/error`, {
          requestId: session.pendingRequest.id,
          message,
        });
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <main style={{ maxWidth: 960, margin: "0 auto", padding: 24, display: "grid", gap: 16 }}>
      <section style={sectionStyle()}>
        <h1 style={{ marginTop: 0 }}>Contracts Deploy UI</h1>
        <p style={{ marginBottom: 0, color: "#cbd5e1" }}>{statusMessage}</p>
      </section>

      {sessionEnded ? (
        <section
          style={{
            ...sectionStyle(),
            borderColor: "#1e3a5f",
            backgroundColor: "#0f1f35",
          }}
        >
          <h2 style={{ marginTop: 0 }}>Session disconnected</h2>
          <p style={{ marginBottom: 0, color: "#cbd5e1" }}>
            The Hardhat deploy bridge has stopped. Polling is disabled. If you are not running another deployment, you
            can close this browser tab.
          </p>
        </section>
      ) : null}

      <section style={sectionStyle()}>
        <h2 style={{ marginTop: 0 }}>Wallet</h2>
        <p>
          {!walletUiReady
            ? "Preparing wallet connection…"
            : isConnected && address
              ? `Connected as ${address}`
              : "Connect the wallet that should sign this deploy file."}
        </p>
        <div style={{ display: "flex", gap: 12 }}>
          {!walletUiReady ? null : !isConnected && injectedConnector ? (
            <button
              style={buttonStyle(isConnecting)}
              disabled={isConnecting}
              onClick={() => connect({ connector: injectedConnector })}
            >
              {isConnecting ? "Connecting..." : "Connect Wallet"}
            </button>
          ) : null}
          {!walletUiReady ? null : isConnected ? (
            <button style={buttonStyle(false)} onClick={() => disconnect()}>
              Disconnect
            </button>
          ) : null}
        </div>
        {connectError ? <p style={{ color: "#fca5a5" }}>{connectError.message}</p> : null}
      </section>

      <section style={sectionStyle()}>
        <h2 style={{ marginTop: 0 }}>Deployment Target</h2>
        <p style={{ marginBottom: 8 }}>Deploy file: {session?.deployFile ?? "Loading..."}</p>
        <p style={{ marginBottom: 8 }}>Network: {session?.networkName ?? "Loading..."}</p>
        <p style={{ marginBottom: 8 }}>Chain: {session?.chain.chainName ?? "Loading..."}</p>
        <p style={{ marginBottom: 0 }}>Chain ID: {session?.chain.chainId ?? "Loading..."}</p>
      </section>

      <section style={sectionStyle()}>
        <h2 style={{ marginTop: 0 }}>Pending Request</h2>
        {session?.pendingRequest ? (
          <>
            <p style={{ marginBottom: 8 }}>
              <strong>{session.pendingRequest.label}</strong>
            </p>
            <p style={{ marginTop: 0 }}>{session.pendingRequest.description}</p>
            {session.pendingRequest.deploymentDetails ? (
              <>
                <h3 style={{ fontSize: 14, marginBottom: 8, color: "#94a3b8" }}>
                  Deployment parameters (from Hardhat)
                </h3>
                <pre style={rawTransactionPreviewStyle()}>
                  {JSON.stringify(session.pendingRequest.deploymentDetails, null, 2)}
                </pre>
              </>
            ) : null}
            <h3 style={{ fontSize: 14, marginBottom: 8, color: "#94a3b8" }}>Raw transaction request</h3>
            <pre style={rawTransactionPreviewStyle()}>{JSON.stringify(session.pendingRequest.request, null, 2)}</pre>
            <button
              style={buttonStyle(isSubmitting || !walletUiReady || !isConnected)}
              disabled={isSubmitting || !walletUiReady || !isConnected}
              onClick={() => void approvePendingRequest()}
            >
              {isSubmitting ? "Submitting..." : "Switch Chain, Sign, and Send"}
            </button>
          </>
        ) : (
          <p style={{ marginBottom: 0 }}>No transaction is waiting for approval yet.</p>
        )}
      </section>

      {actionError ? (
        <section style={{ ...sectionStyle(), borderColor: "#7f1d1d", backgroundColor: "#2a1117" }}>
          <h2 style={{ marginTop: 0 }}>Error</h2>
          <p style={{ marginBottom: 0, color: "#fecaca" }}>{actionError}</p>
        </section>
      ) : null}
    </main>
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
