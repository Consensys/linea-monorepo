import { Page } from "@playwright/test";

export async function getBridgeTransactionsCountImpl(page: Page): Promise<number> {
  const txList = page.getByTestId("native-bridge-transaction-history-list");
  await txList.waitFor({ state: "visible" });
  const txs = txList.getByRole("listitem");
  const txCount = txs.count();
  return txCount;
}
