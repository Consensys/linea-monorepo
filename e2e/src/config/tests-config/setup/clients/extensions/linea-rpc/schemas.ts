import type { LineaCancelBundleRpc } from "./linea-cancel-bundle";
import type { LineaGetProofRpc } from "./linea-get-proof";
import type { GetTransactionExclusionStatusV1Rpc } from "./linea-get-transaction-exclusion-status-v1";
import type { SaveRejectedTransactionV1Rpc } from "./linea-save-rejected-transaction-v1";
import type { LineaSendBundleRpc } from "./linea-send-bundle";
import type { GetZkEVMBlockNumberRpc } from "./rollup-get-zkevm-block-number";
import type { RollupGetZkEVMStateMerkleProofV0Rpc } from "./rollup-get-zkevm-state-merkle-proof-v0";

export type BesuNodeRpcSchema = [LineaSendBundleRpc];

export type SequencerRpcSchema = [LineaCancelBundleRpc];

export type ShomeiRpcSchema = [RollupGetZkEVMStateMerkleProofV0Rpc, GetZkEVMBlockNumberRpc];

export type ShomeiFrontendRpcSchema = [LineaGetProofRpc, GetZkEVMBlockNumberRpc];

export type TransactionExclusionRpcSchema = [GetTransactionExclusionStatusV1Rpc, SaveRejectedTransactionV1Rpc];
