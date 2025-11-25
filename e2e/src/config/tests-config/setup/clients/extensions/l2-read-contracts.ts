import { getContract } from "viem";
import {
  DummyContractAbi,
  TestERC20Abi,
  L2MessageServiceV1Abi,
  TokenBridgeV1_1Abi,
  TestContractAbi,
  SparseMerkleProofAbi,
  LineaSequencerUptimeFeedAbi,
  OpcodeTesterAbi,
} from "../../../../../generated";
import { L2Config } from "../../../config/config-schema";
import { ClientFactory } from "../client-factory";

export function createL2ReadContractsExtension(cfg: L2Config) {
  return (client: ReturnType<ClientFactory["getPublic"]>) => ({
    contracts: {
      dummy: {
        ...getContract({
          abi: DummyContractAbi,
          address: cfg.dummyContractAddress,
          client,
        }).read,
        address: cfg.dummyContractAddress,
      },
      erc20: {
        ...getContract({
          abi: TestERC20Abi,
          address: cfg.l2TokenAddress,
          client,
        }).read,
        address: cfg.l2TokenAddress,
      },
      messageService: {
        ...getContract({
          abi: L2MessageServiceV1Abi,
          address: cfg.l2MessageServiceAddress,
          client,
        }).read,
        address: cfg.l2MessageServiceAddress,
      },
      tokenBridge: {
        ...getContract({
          abi: TokenBridgeV1_1Abi,
          address: cfg.tokenBridgeAddress,
          client,
        }).read,
        address: cfg.tokenBridgeAddress,
      },
      testContract: {
        ...getContract({
          abi: TestContractAbi,
          address: cfg.l2TestContractAddress ?? "0x",
          client,
        }).read,
        address: cfg.l2TestContractAddress ?? "0x",
      },
      sparseMerkleProof: {
        ...getContract({
          abi: SparseMerkleProofAbi,
          address: cfg.l2SparseMerkleProofAddress,
          client,
        }).read,
        address: cfg.l2SparseMerkleProofAddress,
      },
      lineaSequencerUptimeFeed: {
        ...getContract({
          abi: LineaSequencerUptimeFeedAbi,
          address: cfg.l2LineaSequencerUptimeFeedAddress,
          client,
        }).read,
        address: cfg.l2LineaSequencerUptimeFeedAddress,
      },
      opcodeTester: {
        ...getContract({
          abi: OpcodeTesterAbi,
          address: cfg.opcodeTesterAddress,
          client,
        }).read,
        address: cfg.opcodeTesterAddress,
      },
    },
  });
}
