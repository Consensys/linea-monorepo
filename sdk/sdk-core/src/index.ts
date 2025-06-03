export { SparseMerkleTree } from "./merkle-tree/smt";

export { parseBlockExtraData } from "./utils/block";
export { formatMessageStatus } from "./utils/message";
export { getContractsAddressesByChainId } from "./utils/contract";

export type { L1PublicClient, L2PublicClient } from "./types/client/public";
export type { L1WalletClient, L2WalletClient } from "./types/client/wallet";
export type { OnChainMessageStatus, ExtendedMessage, MessageProof, Message } from "./types/message";
