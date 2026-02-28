import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
import { ProxyAdminModule } from "../lib/ProxyModule";

export interface YieldManagerInitParams {
  pauseTypeRoles: Array<{ pauseType: number; role: string }>;
  unpauseTypeRoles: Array<{ pauseType: number; role: string }>;
  roleAddresses: Array<{ addressWithRole: string; role: string }>;
  initialL2YieldRecipients: string[];
  defaultAdmin: string;
  initialMinimumWithdrawalReservePercentageBps: number;
  initialTargetWithdrawalReservePercentageBps: number;
  initialMinimumWithdrawalReserveAmount: bigint;
  initialTargetWithdrawalReserveAmount: bigint;
}

export interface YieldManagerModuleParams {
  lineaRollupAddress: string;
  initParams: YieldManagerInitParams;
  verifierAdmin: string;
  gIFirstValidator: string;
  gIPendingPartialWithdrawalsRoot: string;
  vaultHub: string;
  vaultFactory: string;
  steth: string;
}

const YieldManagerModule = buildModule("YieldManager", (m) => {
  const { proxyAdmin } = m.useModule(ProxyAdminModule);

  const lineaRollupAddress = m.getParameter<string>("lineaRollupAddress");
  const initParams = m.getParameter<YieldManagerInitParams>("initParams");
  const verifierAdmin = m.getParameter<string>("verifierAdmin");
  const gIFirstValidator = m.getParameter<string>("gIFirstValidator");
  const gIPendingPartialWithdrawalsRoot = m.getParameter<string>("gIPendingPartialWithdrawalsRoot");
  const vaultHub = m.getParameter<string>("vaultHub");
  const vaultFactory = m.getParameter<string>("vaultFactory");
  const steth = m.getParameter<string>("steth");

  const implementation = m.contract("YieldManager", [lineaRollupAddress], {
    id: "YieldManagerImplementation",
  });

  const initializeData = m.encodeFunctionCall(implementation, "initialize", [initParams]);

  const proxy = m.contract("TransparentUpgradeableProxy", [implementation, proxyAdmin, initializeData], {
    id: "YieldManagerProxy",
  });

  const yieldManager = m.contractAt("YieldManager", proxy, {
    id: "YieldManager",
  });

  const validatorContainerProofVerifier = m.contract(
    "ValidatorContainerProofVerifier",
    [verifierAdmin, gIFirstValidator, gIPendingPartialWithdrawalsRoot],
    { id: "ValidatorContainerProofVerifier" },
  );

  const lidoStVaultYieldProviderFactory = m.contract(
    "LidoStVaultYieldProviderFactory",
    [lineaRollupAddress, proxy, vaultHub, vaultFactory, steth, validatorContainerProofVerifier],
    {
      id: "LidoStVaultYieldProviderFactory",
      after: [proxy, validatorContainerProofVerifier],
    },
  );

  return {
    proxyAdmin,
    implementation,
    proxy,
    yieldManager,
    validatorContainerProofVerifier,
    lidoStVaultYieldProviderFactory,
  };
});

export default YieldManagerModule;
