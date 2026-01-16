import { Chain, Account, Client, Transport, Hex, BlockNumber, Address, Hash } from "viem";

export type GetTransactionExclusionStatusV1Parameters = {
  txHash: Hash;
};

export type GetTransactionExclusionStatusV1ReturnType = {
  txHash: Hash;
  from: Address;
  nonce: Hex;
  txRejectionStage: "SEQUENCER" | "RPC" | "P2P";
  reasonMessage: string;
  blockNumber: BlockNumber<Hex>;
  timestamp: string;
};

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
        Parameters: [GetTransactionExclusionStatusV1Parameters["txHash"]];
        ReturnType: GetTransactionExclusionStatusV1ReturnType;
      },
    ]
  >,
  params: GetTransactionExclusionStatusV1Parameters,
): Promise<GetTransactionExclusionStatusV1ReturnType> {
  return client.request({
    method: "linea_getTransactionExclusionStatusV1",
    params: [params.txHash],
  });
}
