export { SparseMerkleTree } from "./merkle-tree/smt";

export { DEFAULT_L2_MESSAGE_TREE_DEPTH, MAX_L2_MESSAGE_TREE_DEPTH } from "./constants/message";

export { parseBlockExtraData } from "./utils/block";
export { formatMessageStatus } from "./utils/message";
export { validateL2MessageTreeDepth, validateL2MessageTreeDepthFromLog } from "./utils/message-tree-depth";
export { getContractsAddressesByChainId } from "./utils/contract";
export { isLineaMainnet, isLineaSepolia, isMainnet, isSepolia } from "./utils/chain";

export type { L1PublicClient, L2PublicClient } from "./types/client/public";
export type { L1WalletClient, L2WalletClient } from "./types/client/wallet";
export type { ExtendedMessage, MessageProof, Message } from "./types/message";
export { OnChainMessageStatus } from "./types/message";
