import { PostmanOptions } from "./config/config";
import { getConfig } from "./config/utils";
import { PostmanApp } from "./PostmanApp";

async function main(): Promise<void> {
  // Load options from environment or config file here.
  // This is a placeholder — wire up your config loading before using in production.
  const options: PostmanOptions = JSON.parse(process.env.POSTMAN_CONFIG ?? "{}") as PostmanOptions;

  // Validate config eagerly
  getConfig(options);

  const app = new PostmanApp(options);

  process.on("SIGTERM", async () => {
    await app.stop();
    process.exit(0);
  });

  process.on("SIGINT", async () => {
    await app.stop();
    process.exit(0);
  });

  await app.start();
}

main().catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});
