import type TestSetupCore from "./test-setup-core";
import { createTestSetup } from "./test-setup-factory";

export type TestContext = TestSetupCore;

export function createTestContext(): TestContext {
  return createTestSetup();
}
