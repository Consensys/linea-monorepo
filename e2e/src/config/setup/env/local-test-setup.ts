import { Config } from "../../schema/config-schema";
import TestSetupCore from "../test-setup-core";

export class LocalTestSetup extends TestSetupCore {
  constructor(config: Config) {
    super(config);
  }

  get environmentName(): string {
    return "local";
  }

  isLocal(): boolean {
    return true;
  }
}
