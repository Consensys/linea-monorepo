import * as dotenv from "dotenv";
import { loadConfigFromEnv } from "./src/application/main/config/loadConfigFromEnv.js";
import { NativeYieldCronJobBootstrap } from "./src/application/main/NativeYieldCronJobBootstrap.js";

dotenv.config();

async function main() {
  const options = loadConfigFromEnv();
  const client = new NativeYieldCronJobBootstrap({
    ...options,
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
