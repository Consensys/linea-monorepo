import { metaMaskFixtures } from "@synthetixio/synpress/playwright";
import setup from "./wallet-setup/metamask.setup";
import { Locator, Page } from "@playwright/test";
import { ETH_SYMBOL } from "./constants";
import { Agent } from "http";

export const test = metaMaskFixtures(setup).extend<{
  // Bridge UI Actions
  clickNativeBridgeButton: () => Promise<Locator>;
  openNativeBridgeTransactionHistory: () => Promise<void>;
  closeNativeBridgeTransactionHistory: () => Promise<void>;
  openNativeBridgeFormSettings: () => Promise<void>;
  toggleShowTestNetworksInNativeBridgeForm: () => Promise<void>;
  getBridgeTransactionsCount: () => Promise<number>;
  // Metamask Actions
  connectMetamaskToDapp: () => Promise<void>;
  waitForTransactionToConfirm: () => Promise<void>;
  confirmTransactionAndWaitForInclusion: () => Promise<void>;
  // Composite Bridge UI + Metamask Actions
  bridgeToken: (tokenSymbol: string, amount: string, isETH?: boolean) => Promise<void>;
  waitForTransactionListUpdate: (txCountBeforeUpdate: number) => Promise<boolean>;
}>({
  // Bridge UI Actions
  clickNativeBridgeButton: async ({ page }, use) => {
    await use(async () => {
      const nativeBridgeBtn = page.getByRole("link").filter({ hasText: "Native Bridge" });
      await nativeBridgeBtn.click();
      return nativeBridgeBtn;
    });
  },
  openNativeBridgeTransactionHistory: async ({ page }, use) => {
    await use(async () => {
      const txHistoryIconButton = page.getByTestId("native-bridge-transaction-history-icon");
      await txHistoryIconButton.click();
    });
  },
  closeNativeBridgeTransactionHistory: async ({ page }, use) => {
    await use(async () => {
      await page
        .locator("div")
        .filter({ hasText: /^Transaction History$/ })
        .getByRole("button")
        .click();
    });
  },
  openNativeBridgeFormSettings: async ({ page }, use) => {
    await use(async () => {
      const formSettingsIconButton = page.getByTestId("native-bridge-form-settings-icon");
      await formSettingsIconButton.click();
    });
  },
  toggleShowTestNetworksInNativeBridgeForm: async ({ page }, use) => {
    await use(async () => {
      await page.getByTestId("native-bridge-test-network-toggle").click();
    });
  },
  getBridgeTransactionsCount: async (
    { page, openNativeBridgeTransactionHistory, closeNativeBridgeTransactionHistory },
    use,
  ) => {
    await use(async () => {
      await openNativeBridgeTransactionHistory();
      const txList = page.getByTestId("native-bridge-transaction-history-list");
      await expect(txList).toBeVisible();
      const txs = txList.getByRole("listitem");
      const txCount = txs.count();
      await closeNativeBridgeTransactionHistory();
      return txCount;
    });
  },
  // Metamask Actions
  connectMetamaskToDapp: async ({ page, metamask }, use) => {
    await use(async () => {
      // Click Connect button
      const connectBtn = page.getByRole("button").filter({ hasText: "Connect" }).first();
      await connectBtn.click();

      // Click on 'Metamask' on the wallet dropdown menu
      const metamaskBtnInDropdownList = page.getByRole("button").filter({ hasText: "MetaMask" }).first();
      await metamaskBtnInDropdownList.click();

      await metamask.connectToDapp();
      await metamask.goBackToHomePage();
      await page.bringToFront();
    });
  },
  waitForTransactionToConfirm: async ({ page, metamask }, use) => {
    await use(async () => {
      await metamask.page.bringToFront();
      await metamask.page.reload();
      const activityButton = metamask.page.locator("button", { hasText: "Activity" });
      await activityButton.waitFor();
      await activityButton.click();

      let txCount = await metamask.page
        .locator(metamask.homePage.selectors.activityTab.pendingApprovedTransactions)
        .count();
      while (txCount > 0) {
        txCount = await metamask.page
          .locator(metamask.homePage.selectors.activityTab.pendingApprovedTransactions)
          .count();
      }
    });
  },
  confirmTransactionAndWaitForInclusion: async ({ page, metamask, waitForTransactionToConfirm }, use) => {
    await use(async () => {
      await metamask.confirmTransaction();
      await waitForTransactionToConfirm();
      await page.bringToFront();
    });
  },
  // Composite Bridge UI + Metamask Actions
  bridgeToken: async ({ page, metamask, waitForTransactionToConfirm }, use) => {
    await use(async (tokenSymbol: string, amount: string) => {
      // Wait for 'balance' state
      await selectTokenAndWaitForBalance(tokenSymbol, page);

      // Input amount
      const amountInput = page.getByRole("textbox", { name: "0" });
      await amountInput.fill(amount);

      // Check if there are sufficient funds available
      const insufficientFundsButton = page.getByRole("button", { name: "Insufficient funds" });
      if ((await insufficientFundsButton.count()) > 0)
        throw "Insufficient funds available, please add some funds before running the test";

      // Check if approval required
      const approvalButton = page.getByRole("button", { name: "Approve Token" });
      if ((await approvalButton.count()) > 0) {
        // Do approval flow
        // const tokenType = isETH ? "eth" : "erc";
        // // Check that this amount has been approved
        // if (tokenType === "erc" && submitBtnDisabled === 1) {
        //   //We need to approve the amount first
        //   const approveBtn = await page.waitForSelector(`#approve-btn`);
        //   await approveBtn.click();
        //   await metamask.page.bringToFront();
        //   await metamask.page.reload();
        //   const nextBtn = metamask.page.locator("button", {
        //     hasText: "Next",
        //   });
        //   await nextBtn.waitFor();
        //   await nextBtn.click();
        //   const approveMMBtn = metamask.page.locator("button", { hasText: "Approve" });
        //   await approveMMBtn.waitFor();
        //   await approveMMBtn.click();
        //   await waitForTransactionToConfirm();
        //   await page.bringToFront();
        // }
      }

      // Wait for "Receive amount", otherwise "Confirm and Bridge" button will silently fail
      const receivedAmountField = page.getByTestId("received-amount-text");
      await receivedAmountField.waitFor({ state: "visible" });

      // Click "Bridge" button
      const bridgeButton = page.getByRole("button", { name: "Bridge" });
      await bridgeButton.waitFor();
      await bridgeButton.click();

      // Click "Confirm and Bridge" button
      const confirmAndBridgeButton = page.getByTestId("confirm-and-bridge-btn");
      await expect(confirmAndBridgeButton).toBeVisible();
      await expect(confirmAndBridgeButton).toBeEnabled();
      await confirmAndBridgeButton.click();
    });
  },
  waitForTransactionListUpdate: async ({ getBridgeTransactionsCount }, use) => {
    await use(async (txCountBeforeUpdate: number) => {
      const maxTries = 20;
      let tryCount = 0;
      let listUpdated = false;
      do {
        const newTxCount = await getBridgeTransactionsCount();

        listUpdated = newTxCount !== txCountBeforeUpdate;
        tryCount++;
      } while (!listUpdated && tryCount < maxTries);

      return listUpdated;
    });
  },
});

async function selectTokenAndWaitForBalance(tokenSymbol: string, page: Page) {
  const openModalBtn = page.getByTestId("native-bridge-open-token-list-modal");
  await openModalBtn.click();
  // Wait for ETH amount to not be 0
  const ethBalance = page.getByTestId(`token-details-eth-amount`);
  while ((await ethBalance.textContent()) === "0 ETH") {
    await page.waitForTimeout(500);
  }
  await page.getByTestId(`token-details-${tokenSymbol.toLowerCase()}-btn`).click();
}

export const { expect, describe } = test;
