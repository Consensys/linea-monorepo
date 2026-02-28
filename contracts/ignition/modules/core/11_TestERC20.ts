import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const TestERC20Module = buildModule("TestERC20", (m) => {
  const name = m.getParameter<string>("name", "TestERC20");
  const symbol = m.getParameter<string>("symbol", "TERC20");
  const initialSupply = m.getParameter<number>("initialSupply", 100000);

  const token = m.contract("TestERC20", [name, symbol, initialSupply], {
    id: "TestERC20",
  });

  return { token };
});

export default TestERC20Module;
