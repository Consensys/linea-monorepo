import * as dotenv from "dotenv";
import { transports } from "winston";
import { NativeYieldCronJobClient } from "../src/application/main/NativeYieldCronJobClient";

dotenv.config();

async function main() {
  const client = new NativeYieldCronJobClient({
    yieldOptions: {
      dataSources: {
        l1RpcUrl: process.env.L1_RPC_URL ?? "",
        stakingGraphQLUrl: process.env.STAKING_GRAPHQL_URL ?? "",
        ipfsBaseUrl: process.env.IPFS_BASE_URL ?? "",
      },
      contractAddresses: {
        lineaRollupContractAddress: process.env.LINEA_ROLLUP_ADDRESS ?? "",
        lazyOracleAddress: process.env.LAZY_ORACLE_ADDRESS ?? "",
        yieldManagerAddress: process.env.YIELD_MANAGER_ADDRESS ?? "",
        lidoYieldProviderAddress: process.env.LIDO_YIELD_PROVIDER_ADDRESS ?? "",
        l2YieldRecipientAddress: process.env.L2_YIELD_RECIPIENT ?? "",
      },
    },
    loggerOptions: {
      level: "info",
      transports: [new transports.Console()],
    },
    // apiOptions: {
    //   port: process.env.API_PORT ? parseInt(process.env.API_PORT) : undefined,
    // },
  });
  await client.connectServices();
  client.startAllServices();
}

main()
  .then()
  .catch((error) => {
    console.error("", error);
    process.exit(1);
  });

process.on("SIGINT", () => {
  process.exit(0);
});

process.on("SIGTERM", () => {
  process.exit(0);
});
