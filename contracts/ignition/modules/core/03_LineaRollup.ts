import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const LineaRollupModule = buildModule("LineaRollup", (m) => {
  const initialStateRootHash = m.getParameter<string>("initialStateRootHash");
  const initialL2BlockNumber = m.getParameter<number>("initialL2BlockNumber");
  const genesisTimestamp = m.getParameter<number>("genesisTimestamp");
  const defaultVerifier = m.getParameter<string>("defaultVerifier");
  const rateLimitPeriodInSeconds = m.getParameter<number>("rateLimitPeriodInSeconds");
  const rateLimitAmountInWei = m.getParameter<string>("rateLimitAmountInWei");
  const roleAddresses = m.getParameter<unknown[]>("roleAddresses");
  const pauseTypeRoles = m.getParameter<unknown[]>("pauseTypeRoles");
  const unpauseTypeRoles = m.getParameter<unknown[]>("unpauseTypeRoles");
  const defaultAdmin = m.getParameter<string>("defaultAdmin");
  const multiCallAddress = m.getParameter<string>("multiCallAddress");
  const yieldManagerAddress = m.getParameter<string>("yieldManagerAddress");

  const implementation = m.contract("LineaRollup", [], {
    id: "LineaRollup_Implementation",
  });

  const proxyAdmin = m.contract("src/_testing/integration/ProxyAdmin.sol:ProxyAdmin", [], {
    id: "LineaRollup_ProxyAdmin",
  });

  const initData = m.encodeFunctionCall(implementation, "initialize", [
    {
      initialStateRootHash,
      initialL2BlockNumber,
      genesisTimestamp,
      defaultVerifier,
      rateLimitPeriodInSeconds,
      rateLimitAmountInWei,
      roleAddresses,
      pauseTypeRoles,
      unpauseTypeRoles,
      defaultAdmin,
      shnarfProvider: "0x0000000000000000000000000000000000000000",
    },
    multiCallAddress,
    yieldManagerAddress,
  ]);

  const proxy = m.contract(
    "src/_testing/integration/TransparentUpgradeableProxy.sol:TransparentUpgradeableProxy",
    [implementation, proxyAdmin, initData],
    { id: "LineaRollup_Proxy" },
  );

  return { proxy, proxyAdmin, implementation };
});

export default LineaRollupModule;
