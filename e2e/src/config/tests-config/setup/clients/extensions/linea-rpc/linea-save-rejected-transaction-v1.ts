import { Chain, Account, Client, Transport, Hex } from "viem";

export type SaveRejectedTransactionV1Parameters = {
  txRejectionStage: "SEQUENCER" | "RPC" | "P2P";
  timestamp: string;
  blockNumber: number | null;
  transactionRLP: Hex;
  reasonMessage: string;
  overflows: { module: string; count: number; limit: number }[];
};

export type SaveRejectedTransactionV1ReturnType = boolean;

export async function saveRejectedTransactionV1<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "linea_saveRejectedTransactionV1";
        Parameters: [SaveRejectedTransactionV1Parameters];
        ReturnType: SaveRejectedTransactionV1ReturnType;
      },
    ]
  >,
  params: SaveRejectedTransactionV1Parameters,
): Promise<SaveRejectedTransactionV1ReturnType> {
  return client.request({
    method: "linea_saveRejectedTransactionV1",
    params: [params],
  });
}
