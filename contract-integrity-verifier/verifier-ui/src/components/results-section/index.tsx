"use client";

import { Disclosure } from "@headlessui/react";
import { useVerifierStore } from "@/stores/verifier";
import { Card } from "@/components/ui/card";
import type { ContractVerificationResult, VerificationStatus } from "@consensys/linea-contract-integrity-verifier";
import styles from "./results-section.module.scss";
import clsx from "clsx";

const statusIcons: Record<VerificationStatus, string> = {
  pass: "✓",
  fail: "✗",
  warn: "!",
  skip: "○",
};

const statusLabels: Record<VerificationStatus, string> = {
  pass: "Pass",
  fail: "Fail",
  warn: "Warning",
  skip: "Skipped",
};

function StatusBadge({ status }: { status: VerificationStatus }) {
  return (
    <span className={clsx(styles.badge, styles[status])}>
      <span aria-hidden="true">{statusIcons[status]}</span>
      {statusLabels[status]}
    </span>
  );
}

interface ContractResultCardProps {
  result: ContractVerificationResult;
  verbose: boolean;
}

function ContractResultCard({ result, verbose }: ContractResultCardProps) {
  const { bytecodeResult, abiResult, stateResult, error } = result;

  // Determine overall status
  let overallStatus: VerificationStatus = "pass";
  if (error) {
    overallStatus = "fail";
  } else {
    const statuses = [bytecodeResult?.status, abiResult?.status, stateResult?.status].filter(
      Boolean,
    ) as VerificationStatus[];

    if (statuses.includes("fail")) {
      overallStatus = "fail";
    } else if (statuses.includes("warn")) {
      overallStatus = "warn";
    } else if (statuses.every((s) => s === "skip")) {
      overallStatus = "skip";
    }
  }

  return (
    <Disclosure>
      {({ open }) => (
        <div className={clsx(styles.resultCard, styles[overallStatus])}>
          <Disclosure.Button className={styles.resultHeader}>
            <div className={styles.resultInfo}>
              <h3 className={styles.contractName}>{result.contract.name}</h3>
              <div className={styles.resultMeta}>
                <span className={styles.chain}>{result.contract.chain}</span>
                <code className={styles.address}>{result.contract.address}</code>
              </div>
            </div>
            <div className={styles.resultActions}>
              <StatusBadge status={overallStatus} />
              <svg
                className={clsx(styles.chevron, open && styles.open)}
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                aria-hidden="true"
              >
                <polyline points="6 9 12 15 18 9" />
              </svg>
            </div>
          </Disclosure.Button>

          <Disclosure.Panel className={styles.resultDetails}>
            {error && (
              <div className={styles.errorSection}>
                <p className={styles.errorMessage}>{error}</p>
              </div>
            )}

            {bytecodeResult && (
              <div className={styles.section}>
                <div className={styles.sectionHeader}>
                  <h4>Bytecode Verification</h4>
                  <StatusBadge status={bytecodeResult.status} />
                </div>
                <p className={styles.sectionMessage}>{bytecodeResult.message}</p>
                {verbose && bytecodeResult.matchPercentage !== undefined && (
                  <p className={styles.detail}>Match: {bytecodeResult.matchPercentage.toFixed(2)}%</p>
                )}
                {verbose && bytecodeResult.localBytecodeLength !== undefined && (
                  <p className={styles.detail}>
                    Local bytecode: {bytecodeResult.localBytecodeLength} bytes, Remote bytecode:{" "}
                    {bytecodeResult.remoteBytecodeLength} bytes
                  </p>
                )}
                {verbose && bytecodeResult.immutableDifferences && bytecodeResult.immutableDifferences.length > 0 && (
                  <div className={styles.detail}>
                    <p>Immutable differences:</p>
                    <ul className={styles.selectorList}>
                      {bytecodeResult.immutableDifferences.map((diff, i) => (
                        <li key={i}>
                          Position {diff.position}: <code>{diff.remoteValue}</code>
                          {diff.possibleType && <span> ({diff.possibleType})</span>}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            )}

            {abiResult && (
              <div className={styles.section}>
                <div className={styles.sectionHeader}>
                  <h4>ABI Verification</h4>
                  <StatusBadge status={abiResult.status} />
                </div>
                <p className={styles.sectionMessage}>{abiResult.message}</p>
                {verbose && abiResult.localSelectors && (
                  <p className={styles.detail}>Local selectors: {abiResult.localSelectors.length}</p>
                )}
                {verbose && abiResult.remoteSelectors && (
                  <p className={styles.detail}>Remote selectors: {abiResult.remoteSelectors.length}</p>
                )}
                {abiResult.missingSelectors && abiResult.missingSelectors.length > 0 && (
                  <div className={styles.detail}>
                    <p>Missing selectors:</p>
                    <ul className={styles.selectorList}>
                      {abiResult.missingSelectors.map((s) => (
                        <li key={s}>
                          <code>{s}</code>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
                {verbose && abiResult.extraSelectors && abiResult.extraSelectors.length > 0 && (
                  <div className={styles.detail}>
                    <p>Extra selectors on chain:</p>
                    <ul className={styles.selectorList}>
                      {abiResult.extraSelectors.map((s) => (
                        <li key={s}>
                          <code>{s}</code>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            )}

            {stateResult && (
              <div className={styles.section}>
                <div className={styles.sectionHeader}>
                  <h4>State Verification</h4>
                  <StatusBadge status={stateResult.status} />
                </div>
                <p className={styles.sectionMessage}>{stateResult.message}</p>

                {/* View Calls - always show table but show params only in verbose */}
                {stateResult.viewCallResults && stateResult.viewCallResults.length > 0 && (
                  <div className={styles.subsection}>
                    <h5>View Calls ({stateResult.viewCallResults.length})</h5>
                    <table className={styles.table}>
                      <thead>
                        <tr>
                          <th>Function</th>
                          {verbose && <th>Params</th>}
                          <th>Expected</th>
                          <th>Actual</th>
                          <th>Status</th>
                        </tr>
                      </thead>
                      <tbody>
                        {stateResult.viewCallResults.map((r, i) => (
                          <tr key={i}>
                            <td>
                              <code>{r.function}</code>
                            </td>
                            {verbose && (
                              <td>
                                <code>{r.params?.length ? JSON.stringify(r.params) : "()"}</code>
                              </td>
                            )}
                            <td>
                              <code>{String(r.expected)}</code>
                            </td>
                            <td>
                              <code>{String(r.actual)}</code>
                            </td>
                            <td>
                              <StatusBadge status={r.status} />
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}

                {/* Slot Checks - always show */}
                {stateResult.slotResults && stateResult.slotResults.length > 0 && (
                  <div className={styles.subsection}>
                    <h5>Slot Checks ({stateResult.slotResults.length})</h5>
                    <table className={styles.table}>
                      <thead>
                        <tr>
                          {verbose && <th>Slot</th>}
                          <th>Name</th>
                          <th>Expected</th>
                          <th>Actual</th>
                          <th>Status</th>
                        </tr>
                      </thead>
                      <tbody>
                        {stateResult.slotResults.map((r, i) => (
                          <tr key={i}>
                            {verbose && (
                              <td>
                                <code>{r.slot}</code>
                              </td>
                            )}
                            <td>{r.name}</td>
                            <td>
                              <code>{String(r.expected)}</code>
                            </td>
                            <td>
                              <code>{String(r.actual)}</code>
                            </td>
                            <td>
                              <StatusBadge status={r.status} />
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}

                {/* Storage Paths - always show */}
                {stateResult.storagePathResults && stateResult.storagePathResults.length > 0 && (
                  <div className={styles.subsection}>
                    <h5>Storage Paths ({stateResult.storagePathResults.length})</h5>
                    <table className={styles.table}>
                      <thead>
                        <tr>
                          <th>Path</th>
                          {verbose && <th>Computed Slot</th>}
                          <th>Expected</th>
                          <th>Actual</th>
                          <th>Status</th>
                        </tr>
                      </thead>
                      <tbody>
                        {stateResult.storagePathResults.map((r, i) => (
                          <tr key={i}>
                            <td>
                              <code>{r.path}</code>
                            </td>
                            {verbose && (
                              <td>
                                <code>{r.computedSlot}</code>
                              </td>
                            )}
                            <td>
                              <code>{String(r.expected)}</code>
                            </td>
                            <td>
                              <code>{String(r.actual)}</code>
                            </td>
                            <td>
                              <StatusBadge status={r.status} />
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}

                {/* Namespace Results - show only in verbose */}
                {verbose && stateResult.namespaceResults && stateResult.namespaceResults.length > 0 && (
                  <div className={styles.subsection}>
                    <h5>Namespace Checks ({stateResult.namespaceResults.length})</h5>
                    {stateResult.namespaceResults.map((ns, ni) => (
                      <div key={ni} className={styles.namespaceGroup}>
                        <p className={styles.namespaceId}>
                          <code>{ns.namespaceId}</code> (base slot: <code>{ns.baseSlot}</code>)
                        </p>
                        <table className={styles.table}>
                          <thead>
                            <tr>
                              <th>Name</th>
                              <th>Expected</th>
                              <th>Actual</th>
                              <th>Status</th>
                            </tr>
                          </thead>
                          <tbody>
                            {ns.variables.map((v, vi) => (
                              <tr key={vi}>
                                <td>{v.name}</td>
                                <td>
                                  <code>{String(v.expected)}</code>
                                </td>
                                <td>
                                  <code>{String(v.actual)}</code>
                                </td>
                                <td>
                                  <StatusBadge status={v.status} />
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </Disclosure.Panel>
        </div>
      )}
    </Disclosure>
  );
}

export function ResultsSection() {
  const { results, clearResults, options } = useVerifierStore();

  if (!results) {
    return null;
  }

  const verbose = options.verbose;

  return (
    <Card title="Verification Results">
      <div className={styles.summary}>
        <div className={clsx(styles.summaryItem, styles.pass)}>
          <span className={styles.summaryValue}>{results.passed}</span>
          <span className={styles.summaryLabel}>Passed</span>
        </div>
        <div className={clsx(styles.summaryItem, styles.fail)}>
          <span className={styles.summaryValue}>{results.failed}</span>
          <span className={styles.summaryLabel}>Failed</span>
        </div>
        <div className={clsx(styles.summaryItem, styles.warn)}>
          <span className={styles.summaryValue}>{results.warnings}</span>
          <span className={styles.summaryLabel}>Warnings</span>
        </div>
        <div className={clsx(styles.summaryItem, styles.skip)}>
          <span className={styles.summaryValue}>{results.skipped}</span>
          <span className={styles.summaryLabel}>Skipped</span>
        </div>
      </div>

      {verbose && <p className={styles.verboseHint}>Verbose mode: showing additional details</p>}

      <div className={styles.results}>
        {results.results.map((result, index) => (
          <ContractResultCard key={index} result={result} verbose={verbose} />
        ))}
      </div>

      <button onClick={clearResults} className={styles.clearButton} type="button">
        Clear results
      </button>
    </Card>
  );
}
