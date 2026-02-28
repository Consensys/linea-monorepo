import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const RecoverFundsModule = buildModule("RecoverFunds", (m) => {
  const executorAddress = m.getParameter<string>("executorAddress");

  const implementation = m.contract("RecoverFunds", [], {
    id: "RecoverFunds_Implementation",
  });

  const proxyAdmin = m.contract("src/_testing/integration/ProxyAdmin.sol:ProxyAdmin", [], {
    id: "RecoverFunds_ProxyAdmin",
  });

  const initData = m.encodeFunctionCall(implementation, "initialize", [executorAddress]);

  const proxy = m.contract(
    "src/_testing/integration/TransparentUpgradeableProxy.sol:TransparentUpgradeableProxy",
    [implementation, proxyAdmin, initData],
    { id: "RecoverFunds_Proxy" },
  );

  return { proxy, proxyAdmin, implementation };
});

export default RecoverFundsModule;
