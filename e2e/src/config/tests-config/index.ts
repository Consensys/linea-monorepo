import TestSetup from "./setup";
import localConfig from "./environments/local";
import devConfig from "./environments/dev";
import sepoliaConfig from "./environments/sepolia";

import { Config } from "./types";

const testEnv = process.env.TEST_ENV || "local"; // Default to local environment

let config: Config;

switch (testEnv) {
  case "dev":
    config = devConfig;
    break;
  case "sepolia":
    config = sepoliaConfig;
    break;
  case "local":
  default:
    config = localConfig;
    break;
}
const setup = new TestSetup(config);

export { setup as config };
