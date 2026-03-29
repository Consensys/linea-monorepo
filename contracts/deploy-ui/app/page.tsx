"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState } from "react";
import { flushSync } from "react-dom";
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
    openZeppelinProxyKind?: "transparent" | "uups" | "beacon";
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
  deployScriptOrdinal?: number;
  batchRunActive?: boolean;
  batchTagsSummary?: string | null;
};

/** One wallet-submitted tx for this browser session; persisted so the UI stays useful after the bridge closes. */
type CompletedDeploymentTx = {
  requestId: string;
  completedAt: string;
  label: string;
  description: string;
  txHash: string;
  from: string;
  chainId: number;
  chainName: string;
  blockExplorerUrls: string[];
  request: TransactionPrompt["request"];
  deploymentDetails: TransactionPrompt["deploymentDetails"];
};

function deployUiHistoryStorageKey(apiBaseUrl: string, sessionId: string): string {
  return `deployUiTxHistory:${apiBaseUrl}:${sessionId}`;
}

function loadDeployHistoryFromStorage(key: string): CompletedDeploymentTx[] {
  try {
    const raw = sessionStorage.getItem(key);
    if (!raw) {
      return [];
    }
    const parsed = JSON.parse(raw) as unknown;
    return Array.isArray(parsed) ? (parsed as CompletedDeploymentTx[]) : [];
  } catch {
    return [];
  }
}

function persistDeployHistory(key: string, items: CompletedDeploymentTx[]): void {
  sessionStorage.setItem(key, JSON.stringify(items));
}

function transactionExplorerUrl(blockExplorerUrls: string[], txHash: string): string | null {
  const base = blockExplorerUrls.find((u) => typeof u === "string" && u.length > 0);
  if (!base) {
    return null;
  }
  const trimmed = base.replace(/\/$/, "");
  return `${trimmed}/tx/${txHash}`;
}

/** Stable fragment for in-page links / `:target` (request ids are UUIDs). */
function deployTxFragmentId(requestId: string): string {
  return `deploy-tx-${requestId.replace(/[^a-zA-Z0-9_-]/g, "-")}`;
}

function scrollToHistoryFragment(frag: string): void {
  window.history.replaceState(null, "", `${window.location.pathname}${window.location.search}#${frag}`);
  const run = () => document.getElementById(frag)?.scrollIntoView({ behavior: "smooth", block: "center" });
  run();
  requestAnimationFrame(run);
  window.setTimeout(run, 100);
}

function deploymentKindBadgeLabels(details: TransactionPrompt["deploymentDetails"]): string[] {
  if (!details) {
    return [];
  }
  const labels: string[] = [];
  if (details.openZeppelinProxyKind === "uups") {
    labels.push("UUPS proxy (OpenZeppelin)");
  } else if (details.openZeppelinProxyKind === "transparent") {
    labels.push("Transparent proxy (OpenZeppelin)");
  } else if (details.openZeppelinProxyKind === "beacon") {
    labels.push("Beacon proxy (OpenZeppelin)");
  }
  const name = (details.contractName ?? "").replace(/\s+/g, "");
  if (/proxyadmin/i.test(name)) {
    labels.push("Proxy admin contract");
  }
  return labels;
}

function DeploymentKindBadges({ details }: { details: TransactionPrompt["deploymentDetails"] }) {
  const labels = deploymentKindBadgeLabels(details);
  if (labels.length === 0) {
    return null;
  }
  return (
    <div className="deploy-kind-badges" aria-label="Deployment type">
      {labels.map((label, index) => (
        <span key={`${label}-${index}`} className="deploy-kind-badge">
          {label}
        </span>
      ))}
    </div>
  );
}

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
  const [deployHistory, setDeployHistory] = useState<CompletedDeploymentTx[]>([]);
  const [statusMessage, setStatusMessage] = useState<string>("Waiting for deployment session...");
  const [actionError, setActionError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  /** Wagmi reconnect differs SSR vs first client paint — gate wallet UI until mount. */
  const [walletUiReady, setWalletUiReady] = useState(false);
  const [sessionEnded, setSessionEnded] = useState(false);
  const hadSuccessfulBridgeFetch = useRef(false);
  /** Prevents double submission before React re-renders (wallet popup / network switch). */
  const approveTransactionFlowRef = useRef(false);

  const injectedConnector = connectors[0];

  useEffect(() => {
    setWalletUiReady(true);
  }, []);

  useEffect(() => {
    if (!apiBaseUrl || !session?.sessionId) {
      return;
    }
    const key = deployUiHistoryStorageKey(apiBaseUrl, session.sessionId);
    setDeployHistory(loadDeployHistoryFromStorage(key));
  }, [apiBaseUrl, session?.sessionId]);

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

  const deployRunLooksComplete =
    sessionEnded && deployHistory.length > 0 && session?.pendingRequest === null;

  const heroSubtitle = useMemo(() => {
    if (!sessionEnded) {
      return statusMessage;
    }
    if (deployHistory.length > 0 && session?.pendingRequest === null) {
      return "Deploy run complete. The Hardhat bridge has closed; submitted transactions are saved below and polling has stopped.";
    }
    if (deployHistory.length > 0) {
      return "Session ended. Submitted transactions are preserved below; polling has stopped.";
    }
    return "Deploy session ended (bridge closed). Polling has stopped.";
  }, [sessionEnded, deployHistory.length, session?.pendingRequest, statusMessage]);

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
    if (deployHistory.length === 0 || typeof window === "undefined") {
      return;
    }
    const raw = window.location.hash.slice(1);
    if (!raw.startsWith("deploy-tx-")) {
      return;
    }
    window.setTimeout(() => scrollToHistoryFragment(raw), 0);
  }, [deployHistory.length]);

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

    if (approveTransactionFlowRef.current || isSubmitting) {
      return;
    }
    approveTransactionFlowRef.current = true;
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

      const historyKey = deployUiHistoryStorageKey(apiBaseUrl, session.sessionId);
      const entry: CompletedDeploymentTx = {
        requestId: session.pendingRequest.id,
        completedAt: new Date().toISOString(),
        label: session.pendingRequest.label,
        description: session.pendingRequest.description,
        txHash: hash,
        from: address,
        chainId: session.chain.chainId,
        chainName: session.chain.chainName,
        blockExplorerUrls: session.chain.blockExplorerUrls,
        request: { ...session.pendingRequest.request },
        deploymentDetails: session.pendingRequest.deploymentDetails ?? null,
      };
      flushSync(() => {
        setDeployHistory((prev) => {
          const next = [...prev.filter((x) => x.requestId !== entry.requestId), entry];
          persistDeployHistory(historyKey, next);
          return next;
        });
      });

      const frag = deployTxFragmentId(entry.requestId);
      scrollToHistoryFragment(frag);

      setStatusMessage(`Submitted ${session.pendingRequest.label} — jumped to #${frag} in the list below.`);
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
      approveTransactionFlowRef.current = false;
      setIsSubmitting(false);
    }
  };

  const pending = session?.pendingRequest;
  const signPendingDisabled = isSubmitting || !walletUiReady || !isConnected || !pending;

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
          <p className="deploy-card__subtitle">{heroSubtitle}</p>
        </section>

        {pending && !sessionEnded ? (
          <section className="deploy-card deploy-quick-action" aria-label="Quick sign current transaction">
            <div className="deploy-quick-action__row">
              <div className="deploy-quick-action__text">
                <h2 className="deploy-quick-action__title">Sign current transaction</h2>
                <p className="deploy-quick-action__label">{pending.label}</p>
                <DeploymentKindBadges details={pending.deploymentDetails} />
              </div>
              <button
                type="button"
                className="deploy-btn deploy-btn--primary deploy-quick-action__btn"
                disabled={signPendingDisabled}
                onClick={() => void approvePendingRequest()}
              >
                {isSubmitting ? "Wallet or network in progress…" : "Switch chain, sign, and send"}
              </button>
            </div>
          </section>
        ) : null}

        {sessionEnded ? (
          <section
            className={`deploy-card ${deployRunLooksComplete ? "deploy-card--complete" : "deploy-card--muted"}`}
          >
            <h2>{deployRunLooksComplete ? "Deploy run complete" : "Session disconnected"}</h2>
            <p>
              {deployRunLooksComplete
                ? "Hardhat finished and closed the local HTTP bridge. The Next.js UI usually keeps running so this tab stays loaded. Details below are kept in session storage; use #deploy-tx-… links to jump to each signed transaction."
                : deployHistory.length > 0
                  ? "The bridge stopped before the last poll showed an idle session—check Hardhat for errors. Any transactions you already signed are listed below."
                  : "The Hardhat deploy bridge has stopped. Polling is disabled. If you are not running another deployment, you can close this browser tab."}
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
          {session?.batchRunActive ? (
            <p className="deploy-batch-note">
              <strong>Batch run:</strong>{" "}
              {session.batchTagsSummary
                ? `Hardhat is running multiple deploy scripts for --tags ${session.batchTagsSummary}. This tab stays open until the full run finishes.`
                : "Hardhat is running multiple deploy scripts in one run. This tab stays open until every script completes."}
            </p>
          ) : null}
          <p className="deploy-meta-row">
            <strong>Current deploy file:</strong> {session?.deployFile ?? "Loading…"}
            {session && (session.deployScriptOrdinal ?? 0) > 0 ? (
              <span className="deploy-meta-row--subtle">
                {" "}
                (script {session.deployScriptOrdinal}
                {session.batchRunActive ? " in this batch" : ""})
              </span>
            ) : null}
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
          <p className="deploy-proxy-tip">
            <strong>Upgradeable (OpenZeppelin) deploys:</strong> Hardhat Upgrades stores prior deployments under{" "}
            <code>.openzeppelin/</code>. If an implementation or ProxyAdmin for this chain is already recorded, only the
            missing contracts are deployed — so you may see <em>one</em> wallet prompt for LineaRollup instead of three
            (implementation + ProxyAdmin + proxy). The rollup you use is still the <strong>proxy</strong> address from
            Hardhat logs; fewer prompts does not mean steps were skipped incorrectly.
          </p>
        </section>

        <section className="deploy-card">
          <h2>Pending request</h2>
          {session?.pendingRequest ? (
            <>
              <p className="deploy-pending-label">{session.pendingRequest.label}</p>
              <DeploymentKindBadges details={session.pendingRequest.deploymentDetails} />
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
                  disabled={signPendingDisabled}
                  onClick={() => void approvePendingRequest()}
                >
                  {isSubmitting ? "Wallet or network in progress…" : "Switch chain, sign, and send"}
                </button>
              </div>
            </>
          ) : (
            <p>No transaction is waiting for approval yet.</p>
          )}
        </section>

        {deployHistory.length > 0 ? (
          <section className="deploy-card">
            <h2>Submitted transactions ({deployHistory.length})</h2>
            <p>
              {sessionEnded
                ? "Saved for this deploy session in your browser (session storage). Starting a new Hardhat run opens a new session."
                : "Each approval is appended here. After Hardhat exits, this list stays visible and polling stops."}
            </p>
            <nav className="deploy-history-toc" aria-label="Jump to signed transaction">
              <span className="deploy-history-toc__title">Jump to</span>
              <ul>
                {[...deployHistory].reverse().map((item) => {
                  const frag = deployTxFragmentId(item.requestId);
                  return (
                    <li key={`jump-${item.requestId}`}>
                      <a
                        className="deploy-history-toc__link"
                        href={`#${frag}`}
                        onClick={(e) => {
                          e.preventDefault();
                          scrollToHistoryFragment(frag);
                        }}
                      >
                        {item.label}
                      </a>
                    </li>
                  );
                })}
              </ul>
            </nav>
            <ul className="deploy-history-list">
              {[...deployHistory].reverse().map((item) => {
                const explorerHref = transactionExplorerUrl(item.blockExplorerUrls, item.txHash);
                const frag = deployTxFragmentId(item.requestId);
                return (
                  <li key={item.requestId} id={frag} className="deploy-history-item">
                    <details>
                      <summary>
                        <div className="deploy-history-summary-box">
                          <span className="deploy-history-summary-text">{item.label}</span>
                          <DeploymentKindBadges details={item.deploymentDetails} />
                        </div>
                      </summary>
                      <div className="deploy-history-meta">
                        <div>
                          <strong>Time:</strong> {new Date(item.completedAt).toLocaleString()}
                        </div>
                        <div>
                          <strong>Tx hash:</strong>{" "}
                          {explorerHref ? (
                            <a className="deploy-hash-link" href={explorerHref} target="_blank" rel="noreferrer">
                              {item.txHash}
                            </a>
                          ) : (
                            <code>{item.txHash}</code>
                          )}
                        </div>
                        <div>
                          <strong>From:</strong> <code>{item.from}</code>
                        </div>
                        {item.request.to ? (
                          <div>
                            <strong>To (contract):</strong> <code>{item.request.to}</code>
                          </div>
                        ) : (
                          <div>
                            <strong>Contract create</strong> — use the explorer or receipt for the deployed address.
                          </div>
                        )}
                        <div>
                          <strong>Chain:</strong> {item.chainName} ({item.chainId})
                        </div>
                        {item.description ? (
                          <div>
                            <strong>Note:</strong> {item.description}
                          </div>
                        ) : null}
                        <div className="deploy-bookmark-row">
                          <span>
                            <strong>Bookmark:</strong>{" "}
                            <a
                              className="deploy-hash-link"
                              href={`#${frag}`}
                              onClick={(e) => {
                                e.preventDefault();
                                scrollToHistoryFragment(frag);
                              }}
                            >
                              #{frag}
                            </a>
                          </span>
                          <button
                            type="button"
                            className="deploy-btn deploy-btn--secondary deploy-btn--small"
                            onClick={() => {
                              const url = new URL(window.location.href);
                              url.hash = frag;
                              void navigator.clipboard.writeText(url.toString());
                            }}
                          >
                            Copy page link
                          </button>
                        </div>
                      </div>
                      {item.deploymentDetails ? (
                        <div className="deploy-history-details">
                          <h3>Deployment parameters</h3>
                          <pre className="deploy-code">
                            {JSON.stringify(item.deploymentDetails, null, 2)}
                          </pre>
                        </div>
                      ) : null}
                      <div className="deploy-history-details">
                        <h3>Raw transaction request</h3>
                        <pre className="deploy-code">{JSON.stringify(item.request, null, 2)}</pre>
                      </div>
                    </details>
                  </li>
                );
              })}
            </ul>
          </section>
        ) : null}

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
