import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const BridgedTokenModule = buildModule("BridgedToken", (m) => {
  const bridgedToken = m.contract("BridgedToken", [], {
    id: "BridgedToken",
  });

  const tokenBeacon = m.contract("UpgradeableBeacon", [bridgedToken], {
    id: "BridgedToken_Beacon",
  });

  return { bridgedToken, tokenBeacon };
});

export default BridgedTokenModule;
