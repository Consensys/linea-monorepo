import { configSchema } from "./config.schema.js";
import { toClientConfig } from "./config.js";

export function loadConfigFromEnv(envObj: NodeJS.ProcessEnv = process.env) {
  const parsed = configSchema.safeParse(envObj);
  if (!parsed.success) {
    // pretty-ish error output for CI/boot logs
    console.error("❌ Invalid configuration:");
    console.error(JSON.stringify(parsed.error.format(), null, 2));
    process.exit(1);
  }
  return toClientConfig(parsed.data);
}
