import { lineaCancelBundle, type LineaCancelBundleParameters, type LineaCancelBundleRpc } from "./linea-cancel-bundle";
import { lineaGetProof, type LineaGetProofParameters, type LineaGetProofRpc } from "./linea-get-proof";
import {
  getTransactionExclusionStatusV1,
  type GetTransactionExclusionStatusV1Parameters,
  type GetTransactionExclusionStatusV1Rpc,
} from "./linea-get-transaction-exclusion-status-v1";
import {
  saveRejectedTransactionV1,
  type SaveRejectedTransactionV1Parameters,
  type SaveRejectedTransactionV1Rpc,
} from "./linea-save-rejected-transaction-v1";
import { lineaSendBundle, type LineaSendBundleParameters, type LineaSendBundleRpc } from "./linea-send-bundle";
import {
  pluginsReloadPluginConfig,
  type PluginsReloadPluginConfigParameters,
  type PluginsReloadPluginConfigRpc,
} from "./plugins-reload-plugin-config";
import { getZkEVMBlockNumber, type GetZkEVMBlockNumberRpc } from "./rollup-get-zkevm-block-number";
import {
  rollupGetZkEVMStateMerkleProofV0,
  type RollupGetZkEVMStateMerkleProofV0Parameters,
  type RollupGetZkEVMStateMerkleProofV0Rpc,
} from "./rollup-get-zkevm-state-merkle-proof-v0";
import { withRetryOnBlockNotFound } from "../../../common/utils/viem-retry";

import type { Client, Transport, Chain, Account } from "viem";

type RpcClient<TRpc extends { Method: string; Parameters?: unknown; ReturnType: unknown }> = Client<
  Transport,
  Chain | undefined,
  Account | undefined,
  [TRpc]
>;

export type BesuNodeActions = ReturnType<ReturnType<typeof createBesuNodeExtension>>;
export type SequencerActions = ReturnType<ReturnType<typeof createSequencerExtension>>;
export type ShomeiActions = ReturnType<ReturnType<typeof createShomeiExtension>>;
export type ShomeiFrontendActions = ReturnType<ReturnType<typeof createShomeiFrontendExtension>>;
export type TransactionExclusionActions = ReturnType<ReturnType<typeof createTransactionExclusionExtension>>;
export type RetryActions = ReturnType<ReturnType<typeof createRetryExtension>>;

export function createBesuNodeExtension() {
  return (client: Client) => ({
    lineaSendBundle: (args: LineaSendBundleParameters) =>
      lineaSendBundle(client as RpcClient<LineaSendBundleRpc>, args),
  });
}

export function createSequencerExtension() {
  return (client: Client) => ({
    lineaCancelBundle: (args: LineaCancelBundleParameters) =>
      lineaCancelBundle(client as RpcClient<LineaCancelBundleRpc>, args),
    pluginsReloadPluginConfig: (args: PluginsReloadPluginConfigParameters) =>
      pluginsReloadPluginConfig(client as RpcClient<PluginsReloadPluginConfigRpc>, args),
  });
}

export function createShomeiExtension() {
  return (client: Client) => ({
    rollupGetZkEVMStateMerkleProofV0: (args: RollupGetZkEVMStateMerkleProofV0Parameters) =>
      rollupGetZkEVMStateMerkleProofV0(client as RpcClient<RollupGetZkEVMStateMerkleProofV0Rpc>, args),
    getZkEVMBlockNumber: () => getZkEVMBlockNumber(client as RpcClient<GetZkEVMBlockNumberRpc>),
  });
}

export function createShomeiFrontendExtension() {
  return (client: Client) => ({
    lineaGetProof: (args: LineaGetProofParameters) => lineaGetProof(client as RpcClient<LineaGetProofRpc>, args),
    getZkEVMBlockNumber: () => getZkEVMBlockNumber(client as RpcClient<GetZkEVMBlockNumberRpc>),
  });
}

export function createTransactionExclusionExtension() {
  return (client: Client) => ({
    getTransactionExclusionStatusV1: (args: GetTransactionExclusionStatusV1Parameters) =>
      getTransactionExclusionStatusV1(client as RpcClient<GetTransactionExclusionStatusV1Rpc>, args),
    saveRejectedTransactionV1: (args: SaveRejectedTransactionV1Parameters) =>
      saveRejectedTransactionV1(client as RpcClient<SaveRejectedTransactionV1Rpc>, args),
  });
}

/**
 * Exposes `withRetryOnBlockNotFound` as a client method. Callers pass an action
 * closure receiving the underlying client so the wrapped call retains full typing
 * (e.g. `client.withRetryOnBlockNotFound((c) => c.estimateFeesPerGas())`).
 *
 * The wrapped action is retried only on `BlockNotFoundError`; all other errors
 * propagate immediately. See `common/utils/viem-retry.ts` for details.
 */
export function createRetryExtension() {
  return <TClient extends Client>(client: TClient) => ({
    withRetryOnBlockNotFound<T>(
      fn: (client: TClient) => Promise<T>,
      options?: { pollingIntervalMs?: number; timeoutMs?: number },
    ): Promise<T> {
      return withRetryOnBlockNotFound(() => fn(client), options);
    },
  });
}
