import { testWithSynpress } from "@synthetixio/synpress";
import { test as advancedFixtures } from "../advancedFixtures";
import { TEST_URL, USDC_SYMBOL, USDC_AMOUNT, WEI_AMOUNT, ETH_SYMBOL } from "../constants";

const test = testWithSynpress(advancedFixtures);

const { expect, describe } = test;

describe("L1 > L2 via Native Bridge", () => {
  test("should successfully go to the bridge UI page", async ({ page }) => {
    const pageUrl = page.url();
    expect(pageUrl).toEqual(TEST_URL);
  });

  test("should have 'Native Bridge' button link on homepage", async ({ clickNativeBridgeButton }) => {
    const nativeBridgeBtn = await clickNativeBridgeButton();
    await expect(nativeBridgeBtn).toBeVisible();
  });

  test("should connect MetaMask to dapp correctly", async ({
    page,
    connectMetamaskToDapp,
    clickNativeBridgeButton,
  }) => {
    await clickNativeBridgeButton();
    await connectMetamaskToDapp();
  });

  test("should be able to load the transaction history", async ({
    page,
    connectMetamaskToDapp,
    clickNativeBridgeButton,
    openNativeBridgeTransactionHistory,
  }) => {
    await connectMetamaskToDapp();
    await clickNativeBridgeButton();
    await openNativeBridgeTransactionHistory();

    const txHistoryHeading = page.getByRole("heading").filter({ hasText: "Transaction History" });
    await expect(txHistoryHeading).toBeVisible();
  });

  test("should be able to switch to test networks", async ({
    page,
    connectMetamaskToDapp,
    clickNativeBridgeButton,
    openNativeBridgeFormSettings,
    toggleShowTestNetworksInNativeBridgeForm,
  }) => {
    await connectMetamaskToDapp();
    await clickNativeBridgeButton();
    await openNativeBridgeFormSettings();
    await toggleShowTestNetworksInNativeBridgeForm();

    // Should have Sepolia text visible
    const sepoliaText = page.getByText("Sepolia").first();
    await expect(sepoliaText).toBeVisible();
  });

  test("should be able to bridge ETH from L1 to L2 in testnet", async ({
    page,
    metamask,
    getBridgeTransactionsCount,
    bridgeToken,
    waitForTransactionToConfirm,
    waitForTransactionListUpdate,
    connectMetamaskToDapp,
    clickNativeBridgeButton,
    openNativeBridgeFormSettings,
    toggleShowTestNetworksInNativeBridgeForm,
  }) => {
    await connectMetamaskToDapp();
    await clickNativeBridgeButton();
    await openNativeBridgeFormSettings();
    await toggleShowTestNetworksInNativeBridgeForm();

    // const txnsLengthBefore = await getBridgeTransactionsCount();

    await bridgeToken(ETH_SYMBOL, WEI_AMOUNT);

    // // Wait for transaction to finish
    // await waitForTransactionToConfirm();

    // We check at the end that the transacton list is updated on the bridge UI
    // const listUpdated = await waitForTransactionListUpdate(txnsLengthBefore);
    // Check that that new transaction is added to the list on the UI
    // expect(listUpdated).toBeTruthy();
  });

  test.skip("should be able to bridge USDC from L1 to L2 in testnet", async ({
    page,
    metamask,
    getBridgeTransactionsCount,
    bridgeToken,
    waitForTransactionToConfirm,
    waitForTransactionListUpdate,
  }) => {
    const txnsLengthBefore = await getBridgeTransactionsCount();

    await bridgeToken(USDC_SYMBOL, USDC_AMOUNT);

    await metamask.confirmTransaction();

    // Wait for transaction to finish
    await waitForTransactionToConfirm();

    await page.bringToFront();
    // We check at the end that the transacton list is updated on the bridge UI
    const listUpdated = await waitForTransactionListUpdate(txnsLengthBefore);

    // Check that that new transaction is added to the list on the UI
    expect(listUpdated).toBeTruthy();
  });

  // test.skip("should be able to bridge ERC20 tokens from L1 to L2", async ({
  //   page,
  //   metamask,
  //   getBridgeTransactionsCount,
  //   selectToken,
  //   bridgeToken,
  //   waitForTransactionListUpdate,
  //   waitForTransactionToConfirm,
  // }) => {
  //   const txnsLengthBefore = await getBridgeTransactionsCount();

  //   // Select WETH in the token list (Easiest to get)
  //   await selectToken("WETH");

  //   await bridgeToken(WEI_AMOUNT);
  //   await metamask.confirmTransaction();

  //   // Wait for transaction to finish
  //   await waitForTransactionToConfirm();

  //   await page.bringToFront();
  //   // We check at the end that the transacton list is updated on the bridge UI
  //   const listUpdated = await waitForTransactionListUpdate(txnsLengthBefore);

  //   // Check that that new transaction is added to the list on the UI
  //   expect(listUpdated).toBeTruthy();
  // });
});
