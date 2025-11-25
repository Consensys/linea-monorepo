import { Client, Hash, Address, Hex, Chain, Account, Transport, BlockNumber } from "viem";
import { generateRandomUUIDv4 } from "./random";

export async function lineaSendBundle<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "linea_sendBundle";
        Parameters: [
          {
            txs: Hash[];
            replacementUUID: string;
            blockNumber: string;
          },
        ];
        ReturnType: {
          bundleHash: Hash;
        };
      },
    ]
  >,
  params: {
    txs: Hash[];
    replacementUUID: string;
    blockNumber: string;
  },
) {
  return client.request({
    method: "linea_sendBundle",
    params: [params],
  });
}

export async function isSendBundleMethodNotFound<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<Transport, chain, account>,
  targetBlockNumber = "0xffff",
) {
  try {
    await lineaSendBundle(client, {
      txs: [],
      replacementUUID: generateRandomUUIDv4(),
      blockNumber: targetBlockNumber,
    });
  } catch (err) {
    if (err instanceof Error) {
      if (err.message === "Method not found") {
        return true;
      }
    }
  }
  return false;
}

export async function lineaCancelBundle<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "linea_cancelBundle";
        Parameters: [string];
        ReturnType: boolean;
      },
    ]
  >,
  params: {
    replacementUUID: string;
  },
) {
  return client.request({
    method: "linea_cancelBundle",
    params: [params.replacementUUID],
  });
}

export async function rollupGetZkEVMStateMerkleProofV0<
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "rollup_getZkEVMStateMerkleProofV0";
        Parameters: [
          {
            startBlockNumber: number;
            endBlockNumber: number;
            zkStateManagerVersion: string;
          },
        ];
        ReturnType: {
          zkEndStateRootHash: Hash;
        };
      },
    ]
  >,
  params: {
    startBlockNumber: number;
    endBlockNumber: number;
    zkStateManagerVersion: string;
  },
) {
  return client.request({
    method: "rollup_getZkEVMStateMerkleProofV0",
    params: [params],
  });
}

export async function lineaGetProof<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "linea_getProof";
        Parameters: [string, string[], string];
        ReturnType: {
          accountProof: {
            key: Hex;
            leafIndex: number;
            proof: {
              value: Hex;
              proofRelatedNodes: Hex[];
            };
          };
          storageProofs: {
            key: Hex;
            leafIndex: number;
            proof: {
              value: Hex;
              proofRelatedNodes: Hex[];
            };
          }[];
        };
      },
    ]
  >,
  params: {
    address: string;
    storageKeys: string[];
    blockParameter: string;
  },
) {
  const { address, storageKeys = [], blockParameter = "latest" } = params;

  return client.request({
    method: "linea_getProof",
    params: [address, storageKeys, blockParameter],
  });
}

export async function getTransactionExclusionStatusV1<
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "linea_getTransactionExclusionStatusV1";
        Parameters: [Hash];
        ReturnType: {
          txHash: Hash;
          from: Address;
          nonce: Hex;
          txRejectionStage: "SEQUENCER" | "RPC" | "P2P";
          reasonMessage: string;
          blockNumber: BlockNumber<Hex>;
          timestamp: string;
        };
      },
    ]
  >,
  params: {
    txHash: Hash;
  },
) {
  return client.request({
    method: "linea_getTransactionExclusionStatusV1",
    params: [params.txHash],
  });
}

export async function saveRejectedTransactionV1<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "linea_saveRejectedTransactionV1";
        Parameters: [
          {
            txRejectionStage: "SEQUENCER" | "RPC" | "P2P";
            timestamp: string;
            blockNumber: number | null;
            transactionRLP: Hex;
            reasonMessage: string;
            overflows: { module: string; count: number; limit: number }[];
          },
        ];
        ReturnType: boolean;
      },
    ]
  >,
  params: {
    txRejectionStage: "SEQUENCER" | "RPC" | "P2P";
    timestamp: string;
    blockNumber: number | null;
    transactionRLP: Hex;
    reasonMessage: string;
    overflows: { module: string; count: number; limit: number }[];
  },
) {
  return client.request({
    method: "linea_saveRejectedTransactionV1",
    params: [params],
  });
}

export async function getZkEVMBlockNumber<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "rollup_getZkEVMBlockNumber";
        Parameters: [];
        ReturnType: number;
      },
    ]
  >,
) {
  return client.request({
    method: "rollup_getZkEVMBlockNumber",
    params: [],
  });
}
