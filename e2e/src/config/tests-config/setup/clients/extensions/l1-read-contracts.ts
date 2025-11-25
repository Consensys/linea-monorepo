import { getContract } from "viem";
import { DummyContractAbi, TestERC20Abi, TokenBridgeV1_1Abi, LineaRollupV6Abi } from "../../../../../generated";
import { L1Config } from "../../../config/config-schema";
import { ClientFactory } from "../client-factory";

export function createL1ReadContractsExtension(cfg: L1Config) {
  return (client: ReturnType<ClientFactory["getPublic"]>) => ({
    contracts: {
      rollup: {
        ...getContract({
          abi: LineaRollupV6Abi,
          address: cfg.lineaRollupAddress,
          client,
        }).read,
        address: cfg.lineaRollupAddress,
      },
      proxyAdmin: {
        ...getContract({
          abi: TestERC20Abi,
          address: cfg.lineaRollupProxyAdminAddress,
          client,
        }).read,
        address: cfg.lineaRollupProxyAdminAddress,
      },
      erc20: {
        ...getContract({
          abi: TestERC20Abi,
          address: cfg.l1TokenAddress,
          client,
        }).read,
        address: cfg.l1TokenAddress,
      },
      tokenBridge: {
        ...getContract({
          abi: TokenBridgeV1_1Abi,
          address: cfg.tokenBridgeAddress,
          client,
        }).read,
        address: cfg.tokenBridgeAddress,
      },
      dummy: {
        ...getContract({
          abi: DummyContractAbi,
          address: cfg.dummyContractAddress,
          client,
        }).read,
        address: cfg.dummyContractAddress,
      },
    },
  });
}
