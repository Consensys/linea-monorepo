import { network } from "hardhat";

import type { Fixture } from "@nomicfoundation/hardhat-network-helpers/types";
import type { DefaultChainType } from "hardhat/types/network";

type NumberLike = number | bigint | string;

async function getNetworkHelpers() {
  const connection = await network.getOrCreate();
  return connection.networkHelpers;
}

export async function loadFixture<T>(fixture: Fixture<T, DefaultChainType>): Promise<T> {
  const networkHelpers = await getNetworkHelpers();
  return networkHelpers.loadFixture(fixture);
}

export async function setNonce(address: string, nonce: NumberLike): Promise<void> {
  const networkHelpers = await getNetworkHelpers();
  await networkHelpers.setNonce(address, nonce);
}

export async function setBalance(address: string, balance: NumberLike): Promise<void> {
  const networkHelpers = await getNetworkHelpers();
  await networkHelpers.setBalance(address, balance);
}

export async function setNextBlockTimestamp(timestamp: NumberLike | Date): Promise<void> {
  const networkHelpers = await getNetworkHelpers();
  await networkHelpers.time.setNextBlockTimestamp(timestamp);
}

export async function clearSnapshots(): Promise<void> {
  const networkHelpers = await getNetworkHelpers();
  networkHelpers.clearSnapshots();
}

export const time = {
  async latest(): Promise<number> {
    const networkHelpers = await getNetworkHelpers();
    return networkHelpers.time.latest();
  },
  async increase(amountInSeconds: NumberLike): Promise<number> {
    const networkHelpers = await getNetworkHelpers();
    return networkHelpers.time.increase(amountInSeconds);
  },
  async setNextBlockTimestamp(timestamp: NumberLike | Date): Promise<void> {
    await setNextBlockTimestamp(timestamp);
  },
};
