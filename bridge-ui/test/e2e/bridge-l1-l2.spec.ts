import { testWithSynpress } from "@synthetixio/synpress";
import { test as advancedFixtures } from "../advancedFixtures";
import { SEPOLIA_NETWORK_NAME, TEST_URL, USDC_AMOUNT, WEI_AMOUNT } from "../constants";

const test = testWithSynpress(advancedFixtures);

const { expect, describe } = test;

describe("Bridge L1 > L2", () => {
  test("should set up the UI and metamask correctly", async ({ page, metamask, initUI }) => {
    await initUI(true);

    await page.locator("#wallet-connect-btn").click();
    await page.locator("wui-list-wallet", { hasText: "MetaMask" }).nth(1).click();

    await metamask.connectToDapp();

    await page.bringToFront();
  });

  test("should successfully go to the bridge UI page", async ({ page }) => {
    const pageUrl = page.url();
    expect(pageUrl).toEqual(TEST_URL);
  });

  test("should successfully display the correct heading", async ({ page, initUI }) => {
    await initUI(true);

    const header = "Bridge";
    await page.locator("h2", { hasText: header }).waitFor({ state: "visible" });
  });

  test("metamask should be connected to the right network", async ({ page, metamask, initUI }) => {
    await initUI(true);

    await page.locator("#wallet-connect-btn").click();
    await page.locator("wui-list-wallet", { hasText: "MetaMask" }).nth(1).click();

    await metamask.connectToDapp();

    await page.bringToFront();
    await page
      .locator("#active-chain-name", {
        hasText: SEPOLIA_NETWORK_NAME,
      })
      .waitFor();
  });

  test.skip("should be able to reload the transaction history", async ({ page, metamask, initUI }) => {
    await initUI(true);
    await page.locator("#wallet-connect-btn").click();
    await page.locator("wui-list-wallet", { hasText: "MetaMask" }).nth(1).click();

    await metamask.connectToDapp();

    const reloadHistoryBtn = await page.waitForSelector("#reload-history-btn");
    await reloadHistoryBtn.click();

    const reloadConfirmBtn = await page.waitForSelector("#reload-history-confirm-btn");
    await reloadConfirmBtn.click();

    await page.locator("#transactions-list").locator("ul").nth(1).waitFor({ timeout: 10_000 });
  });

  test("should be able to switch network", async ({ page, metamask, initUI }) => {
    await initUI(true);
    await page.locator("#wallet-connect-btn").click();
    await page.locator("wui-list-wallet", { hasText: "MetaMask" }).nth(1).click();

    await metamask.connectToDapp();

    await page.locator("#chain-select").click();
    await page.locator("#switch-alternative-chain-btn").click();

    await metamask.approveSwitchNetwork();

    await page.bringToFront();

    await page.locator("#active-chain-name").getByText("Linea Sepolia Testnet").waitFor();
  });

  test.skip("should be able to claim funds if available", async ({
    page,
    metamask,
    initUI,
    waitForTransactionToConfirm,
  }) => {
    await initUI(true);
    await page.locator("#wallet-connect-btn").click();
    await page.locator("wui-list-wallet", { hasText: "MetaMask" }).nth(1).click();

    await metamask.connectToDapp();

    // Check if there is a claim button available
    const checkClaimBtn = await page.locator("#claim-funds-btn").all();
    if (checkClaimBtn.length > 0) {
      const claimBtn = page.locator("#claim-funds-btn").nth(1);
      await claimBtn.click();

      await metamask.confirmTransaction();

      await waitForTransactionToConfirm();
    } else {
      console.warn("Claim funds could not be tested since no funds are waiting to be claimed");
    }
  });

  test.skip("should be able to bridge ETH from L1 to L2", async ({
    page,
    metamask,
    getBridgeTransactionsCount,
    sendTokens,
    waitForTransactionToConfirm,
    waitForTransactionListUpdate,
  }) => {
    await page.bringToFront();
    const txnsLengthBefore = await getBridgeTransactionsCount();

    await sendTokens(WEI_AMOUNT, true);
    await metamask.confirmTransaction();

    // Wait for transaction to finish
    await waitForTransactionToConfirm();

    await page.bringToFront();
    // We check at the end that the transacton list is updated on the bridge UI
    const listUpdated = await waitForTransactionListUpdate(txnsLengthBefore);

    // Check that that new transaction is added to the list on the UI
    expect(listUpdated).toBeTruthy();
  });

  test.skip("should be able to bridge USDC from L1 to L2", async ({
    page,
    metamask,
    getBridgeTransactionsCount,
    selectToken,
    sendTokens,
    waitForTransactionToConfirm,
    waitForTransactionListUpdate,
  }) => {
    const txnsLengthBefore = await getBridgeTransactionsCount();

    // Select USDC in the token list
    await selectToken("USDC");

    await sendTokens(USDC_AMOUNT);

    await metamask.confirmTransaction();

    // Wait for transaction to finish
    await waitForTransactionToConfirm();

    await page.bringToFront();
    // We check at the end that the transacton list is updated on the bridge UI
    const listUpdated = await waitForTransactionListUpdate(txnsLengthBefore);

    // Check that that new transaction is added to the list on the UI
    expect(listUpdated).toBeTruthy();
  });

  test.skip("should be able to bridge ERC20 tokens from L1 to L2", async ({
    page,
    metamask,
    getBridgeTransactionsCount,
    selectToken,
    sendTokens,
    waitForTransactionListUpdate,
    waitForTransactionToConfirm,
  }) => {
    const txnsLengthBefore = await getBridgeTransactionsCount();

    // Select WETH in the token list (Easiest to get)
    await selectToken("WETH");

    await sendTokens(WEI_AMOUNT);
    await metamask.confirmTransaction();

    // Wait for transaction to finish
    await waitForTransactionToConfirm();

    await page.bringToFront();
    // We check at the end that the transacton list is updated on the bridge UI
    const listUpdated = await waitForTransactionListUpdate(txnsLengthBefore);

    // Check that that new transaction is added to the list on the UI
    expect(listUpdated).toBeTruthy();
  });
});
