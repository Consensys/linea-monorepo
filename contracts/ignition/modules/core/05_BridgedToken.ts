import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const BridgedTokenModule = buildModule("BridgedToken", (m) => {
  const bridgedTokenImplementation = m.contract("BridgedToken", [], {
    id: "BridgedTokenImplementation",
  });

  const beacon = m.contract("UpgradeableBeacon", [bridgedTokenImplementation], {
    id: "BridgedTokenBeacon",
  });

  return { bridgedTokenImplementation, beacon };
});

export default BridgedTokenModule;
