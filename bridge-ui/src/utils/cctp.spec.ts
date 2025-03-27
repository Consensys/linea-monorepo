import { test } from "@playwright/test";
import { getCctpMessageExpiryBlock } from "./cctp";

const { expect, describe } = test;

describe("getCctpMessageExpiryBlock", () => {
  test("should return undefined for empty byte string", () => {
    const message = "0x";
    const resp = getCctpMessageExpiryBlock(message);
    expect(resp).toBeUndefined();
  });
});
