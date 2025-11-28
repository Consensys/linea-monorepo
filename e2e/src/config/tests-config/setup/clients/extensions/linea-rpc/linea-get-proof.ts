import { Account, Chain, Client, Hex, Transport } from "viem";

export type LineaGetProofParameters = {
  address: string;
  storageKeys: string[];
  blockParameter: string;
};

export type LineaGetProofReturnType = {
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

export async function lineaGetProof<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "linea_getProof";
        Parameters: [
          LineaGetProofParameters["address"],
          LineaGetProofParameters["storageKeys"],
          LineaGetProofParameters["blockParameter"],
        ];
        ReturnType: LineaGetProofReturnType;
      },
    ]
  >,
  params: LineaGetProofParameters,
): Promise<LineaGetProofReturnType> {
  const { address, storageKeys = [], blockParameter = "latest" } = params;

  return client.request({
    method: "linea_getProof",
    params: [address, storageKeys, blockParameter],
  });
}
