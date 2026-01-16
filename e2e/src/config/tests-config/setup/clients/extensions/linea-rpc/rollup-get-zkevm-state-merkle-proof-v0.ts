import { Account, Chain, Client, Hash, Transport } from "viem";

export type RollupGetZkEVMStateMerkleProofV0Parameters = {
  startBlockNumber: number;
  endBlockNumber: number;
  zkStateManagerVersion: string;
};

export type RollupGetZkEVMStateMerkleProofV0ReturnType = {
  zkEndStateRootHash: Hash;
};

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
        Parameters: [RollupGetZkEVMStateMerkleProofV0Parameters];
        ReturnType: RollupGetZkEVMStateMerkleProofV0ReturnType;
      },
    ]
  >,
  params: RollupGetZkEVMStateMerkleProofV0Parameters,
): Promise<RollupGetZkEVMStateMerkleProofV0ReturnType> {
  return client.request({
    method: "rollup_getZkEVMStateMerkleProofV0",
    params: [params],
  });
}
