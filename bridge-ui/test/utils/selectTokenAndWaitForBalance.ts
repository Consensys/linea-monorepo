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
  const tokenBalance = page.getByTestId(`token-details-${tokenSymbol.toLowerCase()}-amount`);
  if ((await tokenBalance.textContent()) === `0 ${tokenSymbol}`) {
    throw `No ${tokenSymbol} balance, please add some funds before running the test`;
  }
  console.log(`Selected token balance: ${await tokenBalance.textContent()}`);
  // Select token
  await page.getByTestId(`token-details-${tokenSymbol.toLowerCase()}-btn`).click();
}
