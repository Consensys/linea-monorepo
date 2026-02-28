import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const CustomBridgedTokenModule = buildModule("CustomBridgedToken", (m) => {
  const name = m.getParameter<string>("name");
  const symbol = m.getParameter<string>("symbol");
  const decimals = m.getParameter<number>("decimals");

  const implementation = m.contract("CustomBridgedToken", [], {
    id: "CustomBridgedToken_Implementation",
  });

  const proxyAdmin = m.contract("src/_testing/integration/ProxyAdmin.sol:ProxyAdmin", [], {
    id: "CustomBridgedToken_ProxyAdmin",
  });

  const initData = m.encodeFunctionCall(implementation, "initialize", [name, symbol, decimals]);

  const proxy = m.contract(
    "src/_testing/integration/TransparentUpgradeableProxy.sol:TransparentUpgradeableProxy",
    [implementation, proxyAdmin, initData],
    { id: "CustomBridgedToken_Proxy" },
  );

  return { proxy, proxyAdmin, implementation };
});

export default CustomBridgedTokenModule;
