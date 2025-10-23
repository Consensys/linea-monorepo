// pnpm --filter @consensys/linea-shared-utils exec tsx scripts/test-beacon-node-api-client.ts

import { BeaconNodeApiClient } from "../src/clients/BeaconNodeApiClient";

async function main() {
  const rpcUrl = process.env.BEACON_NODE_RPC_URL;

  if (!rpcUrl) {
    console.error("Missing required env var: BEACON_NODE_RPC_URL");
    process.exitCode = 1;
    return;
  }

  const client = new BeaconNodeApiClient(rpcUrl);

  console.log(`Fetching pending partial withdrawals from ${rpcUrl}...`);
  try {
    const withdrawals = await client.getPendingPartialWithdrawals();
    console.log(`Received ${withdrawals.length} withdrawals.`);
    if (withdrawals.length > 0) {
      console.log("Sample entry:", withdrawals[0]);
    }
  } catch (err) {
    console.error("BeaconNodeApiClient integration script failed:", err);
    process.exitCode = 1;
  }
}

main();
