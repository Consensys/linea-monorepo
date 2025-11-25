import { getContract } from "viem";
import { DummyContractAbi, TestERC20Abi, TokenBridgeV1_1Abi, LineaRollupV6Abi } from "../../../../../generated";
import { L1Config } from "../../../config/config-schema";
import { ClientFactory } from "../client-factory";

export function createL1WriteContractsExtension(cfg: L1Config) {
  return (client: ReturnType<ClientFactory["getWallet"]>) => ({
    contracts: {
      rollup: {
        ...getContract({
          abi: LineaRollupV6Abi,
          address: cfg.lineaRollupAddress,
          client,
        }).write,
        address: cfg.lineaRollupAddress,
      },
      proxyAdmin: {
        ...getContract({
          abi: TestERC20Abi,
          address: cfg.lineaRollupProxyAdminAddress,
          client,
        }).write,
        address: cfg.lineaRollupProxyAdminAddress,
      },
      erc20: {
        ...getContract({
          abi: TestERC20Abi,
          address: cfg.l1TokenAddress,
          client,
        }).write,
        address: cfg.l1TokenAddress,
      },
      tokenBridge: {
        ...getContract({
          abi: TokenBridgeV1_1Abi,
          address: cfg.tokenBridgeAddress,
          client,
        }).write,
        address: cfg.tokenBridgeAddress,
      },
      dummy: {
        ...getContract({
          abi: DummyContractAbi,
          address: cfg.dummyContractAddress,
          client,
        }).write,
        address: cfg.dummyContractAddress,
      },
    },
  });
}
