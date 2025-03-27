import { Page } from "@playwright/test";

export async function selectTokenAndWaitForBalance(tokenSymbol: string, page: Page) {
  const openModalBtn = page.getByTestId("native-bridge-open-token-list-modal");
  await openModalBtn.click();
  // Wait for API request to retrieve blockchain balance
  // NB: Assumes wallet has >0 ETH
  const ethBalance = page.getByTestId(`token-details-eth-amount`);
  while ((await ethBalance.textContent()) === "0 ETH") {
    await page.waitForTimeout(250);
  }
  // Throw if no token balance
  // TO investigate - This part seems ~1% flaky, have seen at least one occasion where this returned 0 balance with non-zero ETH balance
  // So assumption that USDC balance is retrieved at the same time as ETH balance may be incorrect.
  const tokenBalance = page.getByTestId(`token-details-${tokenSymbol.toLowerCase()}-amount`);
  if ((await tokenBalance.textContent()) === `0 ${tokenSymbol}`) {
    throw `No ${tokenSymbol} balance, please add some funds before running the test`;
  }
  console.log(`Selected token balance: ${await tokenBalance.textContent()}`);
  // Select token
  await page.getByTestId(`token-details-${tokenSymbol.toLowerCase()}-btn`).click();
}
