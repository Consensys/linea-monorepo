import { Config } from "../../schema/config-schema";
import TestSetupCore from "../test-setup-core";

export class DevTestSetup extends TestSetupCore {
  constructor(config: Config) {
    super(config);
  }

  get environmentName(): string {
    return "dev";
  }

  public isLocal(): boolean {
    return false;
  }
}
