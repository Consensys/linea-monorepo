import { expect } from "chai";
import { BaseContract, ContractFactory, TransactionReceipt } from "ethers";

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

export function expectEventDirectFromReceiptData(
  contract: BaseContract,
  transactionReceipt: TransactionReceipt,
  expectedEventName: string,
  expectedEventArgs: undefined[] = [],
  logIndex: number = 0,
) {
  const logSnippet = {
    topics: transactionReceipt?.logs[logIndex].topics as ReadonlyArray<string>,
    data: transactionReceipt!.logs[logIndex].data,
  };

  const event = contract.interface.parseLog(logSnippet);
  expect(event).is.not.null;
  expect(expectedEventName).equal(event!.name);

  // this is cast to array as the readonly is not compatible with deep
  expect(event!.args.toArray()).to.have.deep.members(expectedEventArgs);
}
