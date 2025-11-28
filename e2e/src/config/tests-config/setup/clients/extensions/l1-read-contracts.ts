import { Account, Chain, Client, Transport } from "viem";
import { L1Config } from "../../../config/config-schema";
import {
  getDummyContract,
  getLineaRollupContract,
  getLineaRollupProxyAdminContract,
  getTestERC20Contract,
  getTokenBridgeContract,
} from "../../contracts/contracts";

type PublicActionsL1<
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
> = {
  getLineaRollup: () => ReturnType<typeof getLineaRollupContract<transport, chain, account>>;
  getLineaRollupProxyAdmin: () => ReturnType<typeof getLineaRollupProxyAdminContract<transport, chain, account>>;
  getTestERC20Contract: () => ReturnType<typeof getTestERC20Contract<transport, chain, account>>;
  getTokenBridgeContract: () => ReturnType<typeof getTokenBridgeContract<transport, chain, account>>;
  getDummyContract: () => ReturnType<typeof getDummyContract<transport, chain, account>>;
};

export function createL1ReadContractsExtension(cfg: L1Config) {
  return <
    chain extends Chain | undefined = Chain | undefined,
    account extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, chain, account>,
  ): PublicActionsL1<Transport, chain, account> => ({
    getLineaRollup: () => getLineaRollupContract(client, cfg.lineaRollupAddress),
    getLineaRollupProxyAdmin: () => getLineaRollupProxyAdminContract(client, cfg.lineaRollupProxyAdminAddress),
    getTestERC20Contract: () => getTestERC20Contract(client, cfg.l1TokenAddress),
    getTokenBridgeContract: () => getTokenBridgeContract(client, cfg.tokenBridgeAddress),
    getDummyContract: () => getDummyContract(client, cfg.dummyContractAddress),
  });
}
