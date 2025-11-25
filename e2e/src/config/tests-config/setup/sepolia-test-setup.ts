import { Config } from "../config/config-schema";
import { L2Client } from "./clients/l2-client";
import TestSetupCore from "./test-setup-core";

export class SepoliaTestSetup extends TestSetupCore {
  constructor(config: Config) {
    super(config);

    const l2Client = new L2Client(config.L2);

    this.L2 = {
      client: l2Client,
    };
  }

  get environmentName(): string {
    return "sepolia";
  }

  public isLocal(): boolean {
    return false;
  }
}
