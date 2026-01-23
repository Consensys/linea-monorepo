import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { TestLineaRollup } from "contracts/typechain-types";
import { FinalizationData, ShnarfDataGenerator } from "../../common/types";
import { Contract } from "ethers";

export type AggregatedProofData = {
  finalShnarf: string;
  parentAggregationFinalShnarf: string;
  aggregatedProof: string;
  aggregatedProverVersion: string;
  aggregatedVerifierIndex: number;
  aggregatedProofPublicInput: string;
  dataHashes: string[];
  dataParentHash: string;
  finalStateRootHash: string;
  parentStateRootHash: string;
  parentAggregationLastBlockTimestamp: number;
  lastFinalizedBlockNumber: number;
  finalTimestamp: number;
  finalBlockNumber: number;
  lastFinalizedL1RollingHash: string;
  l1RollingHash: string;
  lastFinalizedL1RollingHashMessageNumber: number;
  l1RollingHashMessageNumber: number;
  finalFtxRollingHash: string;
  parentAggregationFtxRollingHash: string;
  finalFtxNumber: number;
  parentAggregationFtxNumber: number;
  l2MerkleRoots: string[];
  l2MerkleTreesDepth: number;
  l2MessagingBlocksOffsets: string;
  chainID: number;
  baseFee: number;
  coinBase: string;
  l2MessageServiceAddr: string;
  isAllowedCircuitID: number;
  filteredAddresses: string[];
};

export type FinalizeContext = {
  lineaRollup: TestLineaRollup;
  operator: SignerWithAddress;
};

export type FinalizeProofConfig = {
  proofData: AggregatedProofData;
  blobParentShnarfIndex: number;
  shnarfDataGenerator: ShnarfDataGenerator;
  isMultiple: boolean;
};

export type ExpectedCustomError = {
  name: string;
  args?: unknown[];
};

export type FinalizeParams = {
  context: FinalizeContext;
  proofConfig: FinalizeProofConfig;
  overrides?: Partial<FinalizationData>;
};

export type FinalizeCallForwardingParams = {
  callforwarderAddress: string;
  upgradedContract: Contract;
};

export type FailedFinalizeParams = FinalizeParams & { expectedError: ExpectedCustomError };
export type SucceedFinalizeParams = FinalizeParams;
export type SucceedFinalizeParamsCallForwardingProxy = Omit<FinalizeParams, "context"> & {
  context: FinalizeCallForwardingParams;
};
