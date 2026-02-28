import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export interface TimelockParams {
  minDelay: number;
  proposers: string[];
  executors: string[];
  admin: string;
}

const TimelockModule = buildModule("Timelock", (m) => {
  const minDelay = m.getParameter<number>("minDelay", 0);
  const proposers = m.getParameter<string[]>("proposers");
  const executors = m.getParameter<string[]>("executors");
  const admin = m.getParameter<string>("admin");

  const timelock = m.contract("TimeLock", [minDelay, proposers, executors, admin], {
    id: "TimeLock",
  });

  return { timelock };
});

export default TimelockModule;
