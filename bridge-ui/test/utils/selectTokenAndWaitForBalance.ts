import { Page } from "@playwright/test";

export async function selectTokenAndWaitForBalance(tokenSymbol: string, page: Page) {
  const openModalBtn = page.getByTestId("native-bridge-open-token-list-modal");
  await openModalBtn.click();
  // Wait for ETH amount to not be 0
  const ethBalance = page.getByTestId(`token-details-eth-amount`);
  while ((await ethBalance.textContent()) === "0 ETH") {
    await page.waitForTimeout(500);
  }
  await page.getByTestId(`token-details-${tokenSymbol.toLowerCase()}-btn`).click();
}