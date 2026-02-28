import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
import { ProxyAdminModule } from "../lib/ProxyModule";

export interface CustomBridgedTokenParams {
  name: string;
  symbol: string;
  decimals: number;
  bridgeAddress: string;
}

const CustomBridgedTokenModule = buildModule("CustomBridgedToken", (m) => {
  const { proxyAdmin } = m.useModule(ProxyAdminModule);

  const name = m.getParameter<string>("name");
  const symbol = m.getParameter<string>("symbol");
  const decimals = m.getParameter<number>("decimals");
  const bridgeAddress = m.getParameter<string>("bridgeAddress");

  const implementation = m.contract("CustomBridgedToken", [], {
    id: "CustomBridgedTokenImplementation",
  });

  const initializeData = m.encodeFunctionCall(implementation, "initializeV2", [name, symbol, decimals, bridgeAddress]);

  const proxy = m.contract("TransparentUpgradeableProxy", [implementation, proxyAdmin, initializeData], {
    id: "CustomBridgedTokenProxy",
  });

  const customBridgedToken = m.contractAt("CustomBridgedToken", proxy, {
    id: "CustomBridgedToken",
  });

  return { proxyAdmin, implementation, proxy, customBridgedToken };
});

export default CustomBridgedTokenModule;
