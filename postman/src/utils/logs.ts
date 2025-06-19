import { Direction } from "@consensys/linea-sdk";
import { IMessageDBService } from "../core/persistence/IMessageDBService";
import { ContractTransactionResponse } from "ethers";

export async function getStartingBlocksForLogsFetching(
  config: {
    l1LogsFromBlock: number;
    l2LogsFromBlock: number;
  },
  databaseService: IMessageDBService<ContractTransactionResponse>,
): Promise<{
  l1LogsFromBlock: number;
  l2LogsFromBlock: number;
}> {
  const l1LogsFromBlock = Math.max(config.l1LogsFromBlock, 0);
  const l2LogsFromBlock = Math.max(config.l2LogsFromBlock, 0);

  if (l2LogsFromBlock > 0) {
    return { l1LogsFromBlock, l2LogsFromBlock };
  }

  const minBlockNumber = await databaseService.getMinBlockNumber(Direction.L2_TO_L1);

  if (!minBlockNumber) {
    return { l1LogsFromBlock, l2LogsFromBlock: 0 };
  }

  return {
    l1LogsFromBlock,
    l2LogsFromBlock: minBlockNumber,
  };
}
