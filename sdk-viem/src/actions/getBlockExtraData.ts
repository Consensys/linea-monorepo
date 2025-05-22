import { Account, BlockTag, Chain, Client, GetBlockParameters, Hex, hexToNumber, slice, Transport } from "viem";
import { getBlock } from "viem/actions";

export type GetBlockExtraDataReturnType = {
  version: number;
  fixedCost: number;
  variableCost: number;
  ethGasPrice: number;
};

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
  const block = await getBlock(client, parameters as GetBlockParameters<false, blockTag>);

  const version = slice(block.extraData, 0, 1) as Hex;
  const fixedCost = slice(block.extraData, 1, 5) as Hex;
  const variableCost = slice(block.extraData, 5, 9) as Hex;
  const ethGasPrice = slice(block.extraData, 9, 13) as Hex;

  // Original values are in Kwei; convert them back to wei
  const extraData = {
    version: hexToNumber(version),
    fixedCost: hexToNumber(fixedCost) * 1000,
    variableCost: hexToNumber(variableCost) * 1000,
    ethGasPrice: hexToNumber(ethGasPrice) * 1000,
  };

  return extraData;
}
