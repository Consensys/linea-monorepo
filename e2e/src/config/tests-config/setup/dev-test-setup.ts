import { Config } from "../config/config-schema";
import { L2Client } from "./clients/l2-client";
import TestSetupCore from "./test-setup-core";

export class DevTestSetup extends TestSetupCore {
  constructor(config: Config) {
    super(config);

    const l2Client = new L2Client(config.L2, undefined);

    this.L2 = {
      client: l2Client,
    };
  }

  get environmentName(): string {
    return "dev";
  }

  public isLocal(): boolean {
    return false;
  }
}
