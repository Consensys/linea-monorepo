import { Account, Chain, Client, Transport } from "viem";
import { L2Config } from "../../../config/config-schema";
import {
  getDummyContract,
  getL2MessageServiceContract,
  getLineaSequencerUpTimeFeedContract,
  getOpcodeTesterContract,
  getSparseMerkleProofContract,
  getTestContract,
  getTestERC20Contract,
  getTokenBridgeContract,
} from "../../contracts/contracts";
import {
  rollupGetZkEVMStateMerkleProofV0,
  RollupGetZkEVMStateMerkleProofV0Parameters,
} from "./linea-rpc/rollup-get-zkevm-state-merkle-proof-v0";
import {
  lineaCancelBundle,
  LineaCancelBundleParameters,
  LineaCancelBundleReturnType,
} from "./linea-rpc/linea-cancel-bundle";
import { lineaGetProof, LineaGetProofParameters, LineaGetProofReturnType } from "./linea-rpc/linea-get-proof";
import {
  getTransactionExclusionStatusV1,
  GetTransactionExclusionStatusV1Parameters,
  GetTransactionExclusionStatusV1ReturnType,
} from "./linea-rpc/linea-get-transaction-exclusion-status-v1";
import { lineaSendBundle, LineaSendBundleParameters, LineaSendBundleReturnType } from "./linea-rpc/linea-send-bundle";
import { getZkEVMBlockNumber, GetZkEVMBlockNumberReturnType } from "./linea-rpc/rollup-get-zkevm-block-number";

function addLineaRpcEndpoint(client: Client): L2ReadExtension {
  return {
    rollupGetZkEVMStateMerkleProofV0: (args: RollupGetZkEVMStateMerkleProofV0Parameters) =>
      rollupGetZkEVMStateMerkleProofV0(client, args),
    getZkEVMBlockNumber: () => getZkEVMBlockNumber(client),
    lineaCancelBundle: (args: LineaCancelBundleParameters) => lineaCancelBundle(client, args),
    lineaGetProof: (args: LineaGetProofParameters) => lineaGetProof(client, args),
    lineaSendBundle: (args: LineaSendBundleParameters) => lineaSendBundle(client, args),
    getTransactionExclusionStatusV1: (args: GetTransactionExclusionStatusV1Parameters) =>
      getTransactionExclusionStatusV1(client, args),
  };
}

type L2ReadExtension = {
  rollupGetZkEVMStateMerkleProofV0: (args: {
    startBlockNumber: number;
    endBlockNumber: number;
    zkStateManagerVersion: string;
  }) => ReturnType<typeof rollupGetZkEVMStateMerkleProofV0>;
  getZkEVMBlockNumber: () => Promise<GetZkEVMBlockNumberReturnType>;
  lineaCancelBundle: (args: LineaCancelBundleParameters) => Promise<LineaCancelBundleReturnType>;
  lineaGetProof: (args: LineaGetProofParameters) => Promise<LineaGetProofReturnType>;
  lineaSendBundle: (args: LineaSendBundleParameters) => Promise<LineaSendBundleReturnType>;
  getTransactionExclusionStatusV1: (
    args: GetTransactionExclusionStatusV1Parameters,
  ) => Promise<GetTransactionExclusionStatusV1ReturnType>;
};

type PublicActionsL2<
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
> = {
  getDummyContract: () => ReturnType<typeof getDummyContract<transport, chain, account>>;
  getTestERC20Contract: () => ReturnType<typeof getTestERC20Contract<transport, chain, account>>;
  getL2MessageServiceContract: () => ReturnType<typeof getL2MessageServiceContract<transport, chain, account>>;
  getTokenBridgeContract: () => ReturnType<typeof getTokenBridgeContract<transport, chain, account>>;
  getTestContract: () => ReturnType<typeof getTestContract<transport, chain, account>>;
  getSparseMerkleProofContract: () => ReturnType<typeof getSparseMerkleProofContract<transport, chain, account>>;
  getLineaSequencerUptimeFeedContract: () => ReturnType<
    typeof getLineaSequencerUpTimeFeedContract<transport, chain, account>
  >;
  getOpcodeTesterContract: () => ReturnType<typeof getOpcodeTesterContract<transport, chain, account>>;
};

export function createL2ReadContractsExtension(cfg: L2Config) {
  return <
    chain extends Chain | undefined = Chain | undefined,
    account extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, chain, account>,
  ): PublicActionsL2<Transport, chain, account> & L2ReadExtension => ({
    getDummyContract: () => getDummyContract(client, cfg.dummyContractAddress),
    getTestERC20Contract: () => getTestERC20Contract(client, cfg.l2TokenAddress),
    getL2MessageServiceContract: () => getL2MessageServiceContract(client, cfg.l2MessageServiceAddress),
    getTokenBridgeContract: () => getTokenBridgeContract(client, cfg.tokenBridgeAddress),
    getTestContract: () => getTestContract(client, cfg.l2TestContractAddress ?? "0x"),
    getSparseMerkleProofContract: () => getSparseMerkleProofContract(client, cfg.l2SparseMerkleProofAddress),
    getLineaSequencerUptimeFeedContract: () =>
      getLineaSequencerUpTimeFeedContract(client, cfg.l2LineaSequencerUptimeFeedAddress),
    getOpcodeTesterContract: () => getOpcodeTesterContract(client, cfg.opcodeTesterAddress),
    ...addLineaRpcEndpoint(client),
  });
}
