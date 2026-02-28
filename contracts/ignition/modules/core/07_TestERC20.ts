import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export interface TestERC20Params {
  name: string;
  symbol: string;
  initialSupply: bigint;
}

const TestERC20Module = buildModule("TestERC20", (m) => {
  const name = m.getParameter<string>("name", "TestERC20");
  const symbol = m.getParameter<string>("symbol", "TERC20");
  const initialSupply = m.getParameter<bigint>("initialSupply");

  const testERC20 = m.contract("TestERC20", [name, symbol, initialSupply], {
    id: "TestERC20",
  });

  return { testERC20 };
});

export default TestERC20Module;
