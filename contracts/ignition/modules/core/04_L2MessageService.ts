import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const L2MessageServiceModule = buildModule("L2MessageService", (m) => {
  const rateLimitPeriodInSeconds = m.getParameter<number>("rateLimitPeriodInSeconds");
  const rateLimitAmountInWei = m.getParameter<string>("rateLimitAmountInWei");
  const defaultAdmin = m.getParameter<string>("defaultAdmin");
  const roleAddresses = m.getParameter<unknown[]>("roleAddresses");
  const pauseTypeRoles = m.getParameter<unknown[]>("pauseTypeRoles");
  const unpauseTypeRoles = m.getParameter<unknown[]>("unpauseTypeRoles");

  const implementation = m.contract("L2MessageService", [], {
    id: "L2MessageService_Implementation",
  });

  const proxyAdmin = m.contract("src/_testing/integration/ProxyAdmin.sol:ProxyAdmin", [], {
    id: "L2MessageService_ProxyAdmin",
  });

  const initData = m.encodeFunctionCall(implementation, "initialize", [
    rateLimitPeriodInSeconds,
    rateLimitAmountInWei,
    defaultAdmin,
    roleAddresses,
    pauseTypeRoles,
    unpauseTypeRoles,
  ]);

  const proxy = m.contract(
    "src/_testing/integration/TransparentUpgradeableProxy.sol:TransparentUpgradeableProxy",
    [implementation, proxyAdmin, initData],
    { id: "L2MessageService_Proxy" },
  );

  return { proxy, proxyAdmin, implementation };
});

export default L2MessageServiceModule;
