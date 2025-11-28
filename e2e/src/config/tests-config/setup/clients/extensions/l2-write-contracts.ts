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

type WalletActionsL2<
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

export function createL2WriteContractsExtension(cfg: L2Config) {
  return <
    chain extends Chain | undefined = Chain | undefined,
    account extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, chain, account>,
  ): WalletActionsL2<Transport, chain, account> => ({
    getDummyContract: () => getDummyContract(client, cfg.dummyContractAddress),
    getTestERC20Contract: () => getTestERC20Contract(client, cfg.l2TokenAddress),
    getL2MessageServiceContract: () => getL2MessageServiceContract(client, cfg.l2MessageServiceAddress),
    getTokenBridgeContract: () => getTokenBridgeContract(client, cfg.tokenBridgeAddress),
    getTestContract: () => getTestContract(client, cfg.l2TestContractAddress ?? "0x"),
    getSparseMerkleProofContract: () => getSparseMerkleProofContract(client, cfg.l2SparseMerkleProofAddress),
    getLineaSequencerUptimeFeedContract: () =>
      getLineaSequencerUpTimeFeedContract(client, cfg.l2LineaSequencerUptimeFeedAddress),
    getOpcodeTesterContract: () => getOpcodeTesterContract(client, cfg.opcodeTesterAddress),
  });
}
