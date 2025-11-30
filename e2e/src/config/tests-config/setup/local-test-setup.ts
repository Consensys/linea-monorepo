import { Config, LocalL2Config } from "../config/config-schema";
import { L2Client } from "./clients/l2-client";
// import { L2Contracts } from "./contracts/l2-contracts";
import TestSetupCore from "./test-setup-core";

export class LocalTestSetup extends TestSetupCore {
  constructor(config: Config) {
    super(config);

    const localCfg = config.L2 as LocalL2Config;

    const localL2Client = new L2Client(localCfg, localCfg);

    this.L2 = {
      client: localL2Client,
    };
  }

  get environmentName(): string {
    return "local";
  }

  isLocal(): boolean {
    return true;
  }
}
