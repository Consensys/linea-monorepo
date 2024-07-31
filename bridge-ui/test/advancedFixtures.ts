import { metaMaskFixtures } from "@synthetixio/synpress";
import setup from "./wallet-setup/metamask.setup";

export const test = metaMaskFixtures(setup).extend<{
  initUI: (firstInit?: boolean) => Promise<void>;
  waitForTransactionToConfirm: () => Promise<void>;
  getBridgeTransactionsCount: () => Promise<number>;
  sendTokens: (amount: string, isETH?: boolean) => Promise<void>;
  waitForTransactionListUpdate: (txCountBeforeUpdate: number) => Promise<boolean>;
  selectToken: (tokenName: string) => Promise<void>;
}>({
  initUI: async ({ page }, use) => {
    await use(async (firstInit: boolean = false) => {
      const nativeBridgeBtn = await page.waitForSelector("#native-bridge-btn");
      await nativeBridgeBtn.click();

      if (firstInit) {
        const agreeTermsBtn = await page.waitForSelector("#agree-terms-btn");
        await agreeTermsBtn.click();
      }
    });
  },
  waitForTransactionToConfirm: async ({ metamask }, use) => {
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
  getBridgeTransactionsCount: async ({ page }, use) => {
    await use(async () => {
      const transactionsCount = await page.locator("#transactions-list").locator("ul").count();
      return transactionsCount;
    });
  },
  sendTokens: async ({ page, metamask, waitForTransactionToConfirm }, use) => {
    await use(async (amount: string, isETH = false) => {
      const amountInput = await page.waitForSelector("#amount-input");

      // Sending the smallest amount of USDC
      await amountInput.fill(amount);

      // Check if there are funds available
      const approveBtnDisabled = await page.locator("#approve-btn.btn-disabled").count();
      const submitBtnDisabled = await page.locator("#submit-erc-btn.btn-disabled").count();
      if (approveBtnDisabled === 1 && submitBtnDisabled === 1) {
        throw "No funds available, please add some funds before running the test";
      }

      const tokenType = isETH ? "eth" : "erc";

      // Check that this amount has been approved
      if (tokenType === "erc" && submitBtnDisabled === 1) {
        //We need to approve the amount first
        const approveBtn = await page.waitForSelector(`#approve-btn`);
        await approveBtn.click();

        await metamask.page.bringToFront();
        await metamask.page.reload();
        const nextBtn = metamask.page.locator("button", {
          hasText: "Next",
        });
        await nextBtn.waitFor();
        await nextBtn.click();

        const approveMMBtn = metamask.page.locator("button", { hasText: "Approve" });
        await approveMMBtn.waitFor();
        await approveMMBtn.click();

        await waitForTransactionToConfirm();
        await page.bringToFront();
      }
      const submitBtn = await page.waitForSelector(`#submit-${tokenType}-btn`);
      await submitBtn.click();
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
  selectToken: async ({ page }, use) => {
    await use(async (tokenName: string) => {
      const tokenETHBtn = await page.waitForSelector("#token-select-btn");
      await tokenETHBtn.click();
      const tokenUSDCBtn = await page.waitForSelector(`#token-details-${tokenName}-btn`);
      await tokenUSDCBtn.click();
    });
  },
});

export const { expect, describe } = test;
