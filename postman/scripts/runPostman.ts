import * as dotenv from "dotenv";

import { loadConfigFromEnv } from "../src/infrastructure/config/EnvConfigLoader";
import { startPostman } from "../src/main";

dotenv.config();

async function main() {
  const options = loadConfigFromEnv();
  const { stop } = await startPostman(options);

  process.on("SIGINT", () => {
    stop();
    process.exit(0);
  });

  process.on("SIGTERM", () => {
    stop();
    process.exit(0);
  });
}

main()
  .then()
  .catch((error) => {
    console.error("", error);
    process.exit(1);
  });
