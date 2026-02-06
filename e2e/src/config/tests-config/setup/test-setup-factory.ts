import { LocalTestSetup } from "./env/local-test-setup";
import { DevTestSetup } from "./env/dev-test-setup";
import { SepoliaTestSetup } from "./env/sepolia-test-setup";
import { Config } from "../config/config-schema";
import TestSetupCore from "./test-setup-core";
import localConfig from "../config/local";
import devConfig from "../config/dev";
import sepoliaConfig from "../config/sepolia";

const CLASS_MAP: Record<string, new (c: Config) => TestSetupCore> = {
  local: LocalTestSetup,
  dev: DevTestSetup,
  sepolia: SepoliaTestSetup,
};

const CONFIG_MAP: Record<string, Config> = {
  local: localConfig,
  dev: devConfig,
  sepolia: sepoliaConfig,
};

export function createTestSetup(): TestSetupCore {
  const env = process.env.TEST_ENV ?? "local";

  const SetupClass = CLASS_MAP[env];
  const config = CONFIG_MAP[env];

  if (!SetupClass) {
    throw new Error(`Unknown TEST_ENV "${env}". Expected one of: ${Object.keys(CLASS_MAP).join(", ")}`);
  }

  return new SetupClass(config);
}
