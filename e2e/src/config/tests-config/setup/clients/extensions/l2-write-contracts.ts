import { getContract } from "viem";
import {
  DummyContractAbi,
  TestERC20Abi,
  L2MessageServiceV1Abi,
  TokenBridgeV1_1Abi,
  TestContractAbi,
  LineaSequencerUptimeFeedAbi,
  OpcodeTesterAbi,
} from "../../../../../generated";
import { L2Config } from "../../../config/config-schema";
import { ClientFactory } from "../client-factory";

export function createL2WriteContractsExtension(cfg: L2Config) {
  return (client: ReturnType<ClientFactory["getWallet"]>) => {
    return {
      contracts: {
        dummy: {
          ...getContract({
            abi: DummyContractAbi,
            address: cfg.dummyContractAddress,
            client,
          }).write,
          address: cfg.dummyContractAddress,
        },
        erc20: {
          ...getContract({
            abi: TestERC20Abi,
            address: cfg.l2TokenAddress,
            client,
          }).write,
          address: cfg.l2TokenAddress,
        },
        messageService: {
          ...getContract({
            abi: L2MessageServiceV1Abi,
            address: cfg.l2MessageServiceAddress,
            client,
          }).write,
          address: cfg.l2MessageServiceAddress,
        },
        tokenBridge: {
          ...getContract({
            abi: TokenBridgeV1_1Abi,
            address: cfg.tokenBridgeAddress,
            client,
          }).write,
          address: cfg.tokenBridgeAddress,
        },
        testContract: {
          ...getContract({
            abi: TestContractAbi,
            address: cfg.l2TestContractAddress ?? "0x",
            client,
          }).write,
          address: cfg.l2TestContractAddress ?? "0x",
        },
        lineaSequencerUptimeFeed: {
          ...getContract({
            abi: LineaSequencerUptimeFeedAbi,
            address: cfg.l2LineaSequencerUptimeFeedAddress,
            client,
          }).write,
          address: cfg.l2LineaSequencerUptimeFeedAddress,
        },
        opcodeTester: {
          ...getContract({
            abi: OpcodeTesterAbi,
            address: cfg.opcodeTesterAddress,
            client,
          }).write,
          address: cfg.opcodeTesterAddress,
        },
      },
    };
  };
}
