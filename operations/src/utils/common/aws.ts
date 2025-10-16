import {
  CostExplorerClient,
  CostExplorerClientConfigType,
  GetCostAndUsageCommand,
  GetCostAndUsageCommandInput,
  GetCostAndUsageCommandOutput,
} from "@aws-sdk/client-cost-explorer";

export function createAwsCostExplorerClient(config: CostExplorerClientConfigType): CostExplorerClient {
  return new CostExplorerClient(config);
}

export async function getDailyAwsCosts(
  client: CostExplorerClient,
  params: GetCostAndUsageCommandInput,
): Promise<GetCostAndUsageCommandOutput> {
  const command = new GetCostAndUsageCommand(params);
  return await client.send(command);
}
