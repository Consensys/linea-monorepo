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
  // Select token
  await page.getByTestId(`token-details-${tokenSymbol.toLowerCase()}-btn`).click();
}
