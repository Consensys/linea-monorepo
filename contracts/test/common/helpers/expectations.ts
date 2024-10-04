import { expect } from "chai";
import { BaseContract, ContractFactory } from "ethers";

export async function expectRevertWithCustomError<T extends BaseContract | ContractFactory>(
  contract: T,
  asyncCall: Promise<unknown>,
  errorName: string,
  errorArgs: unknown[] = [],
) {
  await expect(asyncCall)
    .to.be.revertedWithCustomError(contract, errorName)
    .withArgs(...errorArgs);
}

export async function expectRevertWithReason(asyncCall: Promise<unknown>, reason: string) {
  await expect(asyncCall).to.be.revertedWith(reason);
}

export async function expectEvent<T extends BaseContract>(
  contract: T,
  asyncCall: Promise<unknown>,
  eventName: string,
  eventArgs: unknown[] = [],
) {
  await expect(asyncCall)
    .to.emit(contract, eventName)
    .withArgs(...eventArgs);
}

export async function expectEvents<T extends BaseContract>(
  contract: T,
  asyncCall: Promise<unknown>,
  events: { name: string; args: unknown[] }[],
) {
  await Promise.all(events.map((event) => expectEvent(contract, asyncCall, event.name, event.args)));
}
