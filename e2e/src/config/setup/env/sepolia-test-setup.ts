import { Config } from "../../schema/config-schema";
import TestSetupCore from "../test-setup-core";

export class SepoliaTestSetup extends TestSetupCore {
  constructor(config: Config) {
    super(config);
  }

  get environmentName(): string {
    return "sepolia";
  }

  public isLocal(): boolean {
    return false;
  }
}
