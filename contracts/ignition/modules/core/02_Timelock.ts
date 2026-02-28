import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const TimelockModule = buildModule("Timelock", (m) => {
  const minDelay = m.getParameter<number>("minDelay");
  const proposers = m.getParameter<string[]>("proposers");
  const executors = m.getParameter<string[]>("executors");
  const adminAddress = m.getParameter<string>("adminAddress");

  const timelock = m.contract("TimeLock", [minDelay, proposers, executors, adminAddress], {
    id: "TimeLock",
  });

  return { timelock };
});

export default TimelockModule;
