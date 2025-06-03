import { parseBlockExtraData } from "@consensys/linea-sdk-core";
import { Account, BaseError, BlockTag, Chain, Client, GetBlockParameters, Prettify, Transport } from "viem";
import { getBlock } from "viem/actions";
import { linea, lineaSepolia } from "viem/chains";

export type GetBlockExtraDataReturnType = Prettify<{
  version: number;
  fixedCost: number;
  variableCost: number;
  ethGasPrice: number;
}>;

export type GetBlockExtraDataParameters<blockTag extends BlockTag = "latest"> = Omit<
  GetBlockParameters<false, blockTag>,
  "includeTransactions"
>;

export async function getBlockExtraData<
  chain extends Chain | undefined,
  account extends Account | undefined,
  blockTag extends BlockTag = "latest",
>(
  client: Client<Transport, chain, account>,
  parameters: GetBlockExtraDataParameters<blockTag>,
): Promise<GetBlockExtraDataReturnType> {
  if (client.chain?.id !== linea.id && client.chain?.id !== lineaSepolia.id) {
    throw new BaseError("Client chain is not Linea or Linea Sepolia");
  }

  const block = await getBlock(client, parameters as GetBlockParameters<false, blockTag>);

  return parseBlockExtraData(block.extraData);
}
