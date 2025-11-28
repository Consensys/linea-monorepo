import { Account, Address, Chain, Client, getContract, Transport } from "viem";
import {
  DummyContractAbi,
  L2MessageServiceV1Abi,
  LineaRollupV6Abi,
  LineaSequencerUptimeFeedAbi,
  OpcodeTesterAbi,
  ProxyAdminAbi,
  SparseMerkleProofAbi,
  TestContractAbi,
  TestERC20Abi,
  TokenBridgeV1_1Abi,
} from "../../../../generated";

export const getLineaRollupContract = <
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<transport, chain, account>,
  address: Address,
) => {
  return getContract({
    abi: LineaRollupV6Abi,
    address,
    client,
  });
};

export const getLineaRollupProxyAdminContract = <
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<transport, chain, account>,
  address: Address,
) => {
  return getContract({
    abi: ProxyAdminAbi,
    address,
    client,
  });
};

export const getTestERC20Contract = <
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<transport, chain, account>,
  address: Address,
) => {
  return getContract({
    abi: TestERC20Abi,
    address,
    client,
  });
};

export const getTokenBridgeContract = <
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<transport, chain, account>,
  address: Address,
) => {
  return getContract({
    abi: TokenBridgeV1_1Abi,
    address,
    client,
  });
};

export const getDummyContract = <
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<transport, chain, account>,
  address: Address,
) => {
  return getContract({
    abi: DummyContractAbi,
    address,
    client,
  });
};

export const getL2MessageServiceContract = <
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<transport, chain, account>,
  address: Address,
) => {
  return getContract({
    abi: L2MessageServiceV1Abi,
    address,
    client,
  });
};

export const getSparseMerkleProofContract = <
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<transport, chain, account>,
  address: Address,
) => {
  return getContract({
    abi: SparseMerkleProofAbi,
    address,
    client,
  });
};

export const getLineaSequencerUpTimeFeedContract = <
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<transport, chain, account>,
  address: Address,
) => {
  return getContract({
    abi: LineaSequencerUptimeFeedAbi,
    address,
    client,
  });
};

export const getOpcodeTesterContract = <
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<transport, chain, account>,
  address: Address,
) => {
  return getContract({
    abi: OpcodeTesterAbi,
    address,
    client,
  });
};

export const getTestContract = <
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<transport, chain, account>,
  address: Address,
) => {
  return getContract({
    abi: TestContractAbi,
    address,
    client,
  });
};

export const getBridgedTokenContract = <
  transport extends Transport,
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<transport, chain, account>,
  address: Address,
) => {
  return getContract({
    abi: TestERC20Abi,
    address,
    client,
  });
};
