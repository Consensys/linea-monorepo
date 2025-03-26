import { testWithSynpress } from "@synthetixio/synpress";
import { test as advancedFixtures } from "../advancedFixtures";
import { TEST_URL, USDC_AMOUNT, WEI_AMOUNT } from "../constants";

const test = testWithSynpress(advancedFixtures);

const { expect, describe } = test;

describe("L1 > L2 via Native Bridge", () => {
  // test.beforeEach(async ( {context} ) => {
  //   // https://playwright.dev/docs/network#modify-requests
  //   // Circumvent cors error in CI workflow
  //   await context.route("https://app.dynamicauth.com/api/**", async route => {
  //     console.log("intercepted", route.request().url());
  //     const response = await route.fetch();
  //   // Add a prefix to the title.
  //     await route.fulfill({
  //       response,
  //       headers: {
  //         ...response.headers(),
  //         "Access-Control-Allow-Origin": "http://localhost:3000"
  //       }
  //     });
  //   });
  // });

  // test("should successfully go to the bridge UI page", async ({ page }) => {
  //   test.setTimeout(10_000);
  //   const pageUrl = page.url();
  //   expect(pageUrl).toEqual(TEST_URL);
  // });

  // test("should have 'Native Bridge' button link on homepage", async ({ clickNativeBridgeButton }) => {
  //   test.setTimeout(10_000);
  //   const nativeBridgeBtn = await clickNativeBridgeButton();
  //   await expect(nativeBridgeBtn).toBeVisible();
  // });

  test("should connect MetaMask to dapp correctly", async ({ page, connectMetamaskToDapp, clickNativeBridgeButton }) => {
    test.setTimeout(30_000);
    await clickNativeBridgeButton();
    await connectMetamaskToDapp();
  });

  // test("should be able to load the transaction history", async ({
  //   page,
  //   connectMetamaskToDapp,
  //   clickNativeBridgeButton,
  //   openNativeBridgeTransactionHistory,
  // }) => {
  //   await connectMetamaskToDapp();
  //   await clickNativeBridgeButton();
  //   await openNativeBridgeTransactionHistory();

  //   const txHistoryHeading = page.getByRole("heading").filter({ hasText: "Transaction History" });
  //   await expect(txHistoryHeading).toBeVisible();
  // });

  // test("should be able to switch to test networks", async ({
  //   page,
  //   connectMetamaskToDapp,
  //   clickNativeBridgeButton,
  //   openNativeBridgeFormSettings,
  //   toggleShowTestNetworksInNativeBridgeForm,
  // }) => {
  //   await connectMetamaskToDapp();
  //   await clickNativeBridgeButton();
  //   await openNativeBridgeFormSettings();
  //   await toggleShowTestNetworksInNativeBridgeForm();

  //   // Should have Sepolia text visible
  //   const sepoliaText = page.getByText("Sepolia").first();
  //   await expect(sepoliaText).toBeVisible();
  // });

  test.skip("should be able to bridge ETH from L1 to L2 in testnet", async ({
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

  test.skip("should be able to bridge USDC from L1 to L2 in testnet", async ({
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
