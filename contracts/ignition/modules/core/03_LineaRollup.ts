import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
import { ProxyAdminModule } from "../lib/ProxyModule";

export interface LineaRollupInitParams {
  initialStateRootHash: string;
  initialL2BlockNumber: number;
  genesisTimestamp: number;
  defaultVerifier: string;
  rateLimitPeriodInSeconds: number;
  rateLimitAmountInWei: string;
  roleAddresses: Array<{ addressWithRole: string; role: string }>;
  pauseTypeRoles: Array<{ pauseType: number; role: string }>;
  unpauseTypeRoles: Array<{ pauseType: number; role: string }>;
  defaultAdmin: string;
  shnarfProvider: string;
}

export interface LineaRollupModuleParams {
  initParams: LineaRollupInitParams;
  multiCallAddress: string;
  yieldManagerAddress: string;
}

const LineaRollupModule = buildModule("LineaRollup", (m) => {
  const { proxyAdmin } = m.useModule(ProxyAdminModule);

  const initParams = m.getParameter<LineaRollupInitParams>("initParams");
  const multiCallAddress = m.getParameter<string>("multiCallAddress", "0xcA11bde05977b3631167028862bE2a173976CA11");
  const yieldManagerAddress = m.getParameter<string>("yieldManagerAddress");

  const implementation = m.contract("LineaRollup", [], {
    id: "LineaRollupImplementation",
  });

  const initializeData = m.encodeFunctionCall(implementation, "initialize", [
    initParams,
    multiCallAddress,
    yieldManagerAddress,
  ]);

  const proxy = m.contract("TransparentUpgradeableProxy", [implementation, proxyAdmin, initializeData], {
    id: "LineaRollupProxy",
  });

  const lineaRollup = m.contractAt("LineaRollup", proxy, {
    id: "LineaRollup",
  });

  return { proxyAdmin, implementation, proxy, lineaRollup };
});

export default LineaRollupModule;
