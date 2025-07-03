import { testWithSynpress } from "@synthetixio/synpress";
import { test as advancedFixtures } from "../advancedFixtures";
import { TEST_URL, USDC_SYMBOL, USDC_AMOUNT, WEI_AMOUNT, ETH_SYMBOL } from "../constants";

const test = testWithSynpress(advancedFixtures);

const { expect, describe } = test;

// There are known lines causing flaky E2E tests in this test suite, these are annotated by 'bridge-ui-known-flaky-line'
describe("L1 > L2 via Native Bridge", () => {
  describe("No blockchain tx cases", () => {
    test.describe.configure({ mode: "parallel" });

    test("should successfully go to the bridge UI page", async ({ page }) => {
      const pageUrl = page.url();
      expect(pageUrl).toEqual(TEST_URL);
    });

    test("should have 'Native Bridge' button link on homepage", async ({
      clickNativeBridgeButton,
      clickFirstVisitModalConfirmButton,
    }) => {
      const nativeBridgeBtn = await clickNativeBridgeButton();
      await clickFirstVisitModalConfirmButton();
      await expect(nativeBridgeBtn).toBeVisible();
    });

    test("should connect MetaMask to dapp correctly", async ({
      connectMetamaskToDapp,
      clickNativeBridgeButton,
      clickFirstVisitModalConfirmButton,
    }) => {
      await clickNativeBridgeButton();
      await clickFirstVisitModalConfirmButton();
      await connectMetamaskToDapp();
    });

    test("should be able to load the transaction history", async ({
      page,
      connectMetamaskToDapp,
      clickNativeBridgeButton,
      openNativeBridgeTransactionHistory,
      clickFirstVisitModalConfirmButton,
    }) => {
      await connectMetamaskToDapp();
      await clickNativeBridgeButton();
      await clickFirstVisitModalConfirmButton();
      await openNativeBridgeTransactionHistory();

      const txHistoryHeading = page.getByRole("heading").filter({ hasText: "Transaction History" });
      await expect(txHistoryHeading).toBeVisible();
    });

    test("should not be able to bridge on the wrong network", async ({
      page,
      metamask,
      connectMetamaskToDapp,
      clickNativeBridgeButton,
      openNativeBridgeFormSettings,
      selectTokenAndInputAmount,
      swapChain,
      clickFirstVisitModalConfirmButton,
    }) => {
      test.setTimeout(60_000);
      await connectMetamaskToDapp();
      await clickNativeBridgeButton();
      await clickFirstVisitModalConfirmButton();
      await openNativeBridgeFormSettings();

      await swapChain();
      await selectTokenAndInputAmount(ETH_SYMBOL, WEI_AMOUNT);

      // Should have 'Switch to Sepolia' network button visible and enabled
      const switchBtn = page.getByTestId("swap-chain-button");
      await expect(switchBtn).toBeVisible();
      await expect(switchBtn).toBeEnabled();

      // Do network switch
      await switchBtn.click();
      await metamask.approveSwitchNetwork();

      // After network switch, should have 'Bridge' button visible and enabled
      const bridgeButton = page.getByRole("button", { name: "Bridge", exact: true });
      await expect(bridgeButton).toBeVisible();
      await expect(bridgeButton).toBeEnabled();
    });

    test("should see Free gas fees for ETH transfer to L2", async ({
      page,
      connectMetamaskToDapp,
      clickNativeBridgeButton,
      openNativeBridgeFormSettings,
      selectTokenAndInputAmount,
      openGasFeeModal,
      clickFirstVisitModalConfirmButton,
    }) => {
      test.setTimeout(60_000);

      await connectMetamaskToDapp();
      await clickNativeBridgeButton();
      await clickFirstVisitModalConfirmButton();
      await openNativeBridgeFormSettings();

      await selectTokenAndInputAmount(ETH_SYMBOL, WEI_AMOUNT);
      await openGasFeeModal();

      // Assert text items
      const l2NetworkFeeText = page.getByText("L2 fee");
      const freeText = page.getByText("Free");
      await expect(l2NetworkFeeText).toBeVisible();
      await expect(freeText).toBeVisible();
      const listItem = page
        .locator("li")
        .filter({
          has: l2NetworkFeeText,
        })
        .filter({
          has: freeText,
        });
      await expect(listItem).toBeVisible();
    });

    // This test is skipped because CCTP is not supported on the local stack.
    test.skip("should not see Free gas fees for USDC transfer to L2", async ({
      page,
      connectMetamaskToDapp,
      clickNativeBridgeButton,
      openNativeBridgeFormSettings,
      toggleShowTestNetworksInNativeBridgeForm,
      selectTokenAndInputAmount,
      openGasFeeModal,
      clickFirstVisitModalConfirmButton,
    }) => {
      test.setTimeout(60_000);

      await connectMetamaskToDapp();
      await clickNativeBridgeButton();
      await clickFirstVisitModalConfirmButton();
      await openNativeBridgeFormSettings();
      await toggleShowTestNetworksInNativeBridgeForm();

      await selectTokenAndInputAmount(USDC_SYMBOL, USDC_AMOUNT);
      await openGasFeeModal();

      // Assert text items
      const freeText = page.getByText("Free");
      await expect(freeText).not.toBeVisible();
    });

    test("should not see Free gas fees for ETH transfer to L1", async ({
      page,
      connectMetamaskToDapp,
      clickNativeBridgeButton,
      openNativeBridgeFormSettings,
      selectTokenAndInputAmount,
      openGasFeeModal,
      switchToL2Network,
      clickFirstVisitModalConfirmButton,
    }) => {
      test.setTimeout(60_000);

      await connectMetamaskToDapp();
      await clickNativeBridgeButton();
      await clickFirstVisitModalConfirmButton();
      await openNativeBridgeFormSettings();

      await switchToL2Network();
      await selectTokenAndInputAmount(ETH_SYMBOL, WEI_AMOUNT);
      await openGasFeeModal();

      // Assert text items
      const freeText = page.getByText("Free");
      await expect(freeText).not.toBeVisible();
    });
  });

  describe("Blockchain tx cases", () => {
    // If not serial risk colliding nonces -> transactions cancelling each other out
    test.describe.configure({ retries: 1, timeout: 120_000, mode: "serial" });

    test("should be able to initiate bridging ETH from L1 to L2", async ({
      getNativeBridgeTransactionsCount,
      waitForNewTxAdditionToTxList,
      connectMetamaskToDapp,
      clickNativeBridgeButton,
      openNativeBridgeFormSettings,
      selectTokenAndInputAmount,
      doInitiateBridgeTransaction,
      openNativeBridgeTransactionHistory,
      closeNativeBridgeTransactionHistory,
      clickFirstVisitModalConfirmButton,
    }) => {
      // Setup testnet UI
      await connectMetamaskToDapp();
      await clickNativeBridgeButton();
      await clickFirstVisitModalConfirmButton();
      await openNativeBridgeFormSettings();

      // Get # of txs in txHistory before doing bridge tx, so that we can later confirm that our bridge tx shows up in the txHistory.
      await openNativeBridgeTransactionHistory();
      const txnsLengthBefore = await getNativeBridgeTransactionsCount();
      await closeNativeBridgeTransactionHistory();

      // // Actual bridging actions
      await selectTokenAndInputAmount(ETH_SYMBOL, WEI_AMOUNT);
      await doInitiateBridgeTransaction();

      // Check that our bridge tx shows up in the tx history
      await waitForNewTxAdditionToTxList(txnsLengthBefore);
    });

    // This test is skipped because CCTP is not supported on the local stack.
    test.skip("should be able to initiate bridging USDC from L1 to L2 in testnet", async ({
      getNativeBridgeTransactionsCount,
      waitForNewTxAdditionToTxList,
      connectMetamaskToDapp,
      clickNativeBridgeButton,
      openNativeBridgeFormSettings,
      toggleShowTestNetworksInNativeBridgeForm,
      selectTokenAndInputAmount,
      doInitiateBridgeTransaction,
      openNativeBridgeTransactionHistory,
      closeNativeBridgeTransactionHistory,
      doTokenApprovalIfNeeded,
      clickFirstVisitModalConfirmButton,
    }) => {
      // Setup testnet UI
      await connectMetamaskToDapp();
      await clickNativeBridgeButton();
      await clickFirstVisitModalConfirmButton();
      await openNativeBridgeFormSettings();
      await toggleShowTestNetworksInNativeBridgeForm();

      // Get # of txs in txHistory before doing bridge tx, so that we can later confirm that our bridge tx shows up in the txHistory.
      await openNativeBridgeTransactionHistory();
      const txnsLengthBefore = await getNativeBridgeTransactionsCount();
      await closeNativeBridgeTransactionHistory();

      // Actual bridging actions
      await selectTokenAndInputAmount(USDC_SYMBOL, USDC_AMOUNT);
      await doTokenApprovalIfNeeded();
      await doInitiateBridgeTransaction();

      // Check that our bridge tx shows up in the tx history
      await waitForNewTxAdditionToTxList(txnsLengthBefore);
    });

    test("should be able to claim if available READY_TO_CLAIM transactions", async ({
      page,
      connectMetamaskToDapp,
      clickNativeBridgeButton,
      openNativeBridgeFormSettings,
      openNativeBridgeTransactionHistory,
      getNativeBridgeTransactionsCount,
      switchToL2Network,
      doClaimTransaction,
      waitForTxListUpdateForClaimTx,
      clickFirstVisitModalConfirmButton,
    }) => {
      await connectMetamaskToDapp();
      await clickNativeBridgeButton();
      await clickFirstVisitModalConfirmButton();
      await openNativeBridgeFormSettings();

      // Switch to L2 network
      await switchToL2Network();

      // Load tx history
      await openNativeBridgeTransactionHistory();
      await getNativeBridgeTransactionsCount();

      // Find and click READY_TO_CLAIM TX
      const readyToClaimTx = page.getByRole("listitem").filter({ hasText: "Ready to claim" });
      const readyToClaimCount = await readyToClaimTx.count();
      if (readyToClaimCount === 0) return;
      await readyToClaimTx.first().click();

      await doClaimTransaction();

      // Check that tx history has updated accordingly
      await waitForTxListUpdateForClaimTx(readyToClaimCount);
    });
  });
});
