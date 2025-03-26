import { LINEA_SEPOLIA_NETWORK, METAMASK_PASSWORD, METAMASK_SEED_PHRASE, TEST_PRIVATE_KEY } from "../constants";
import { defineWalletSetup } from "@synthetixio/synpress";
import { MetaMask, getExtensionId } from "@synthetixio/synpress/playwright";

export default defineWalletSetup(METAMASK_PASSWORD, async (context, walletPage) => {
  // https://playwright.dev/docs/network#modify-requests
  // Circumvent cors error in CI workflow
  await context.route("**/app.dynamicauth.com/api/**", async route => {
    console.log("Intercepted route:", route.request().url());
    await route.continue({ headers: {
      ...route.request().headers(), 
      "Origin": "http://localhost:3000",
      "Sec-Fetch-Site": "cross-site",
    }});
  });

  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  //@ts-ignore
  const extensionId = await getExtensionId(context, "MetaMask");
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  //@ts-ignore
  const metamask = new MetaMask(context, walletPage, METAMASK_PASSWORD, extensionId);
  await metamask.importWallet(METAMASK_SEED_PHRASE);
  await metamask.importWalletFromPrivateKey(TEST_PRIVATE_KEY);

  await metamask.openSettings();

  const SidebarMenus = metamask.homePage.selectors.settings.SettingsSidebarMenus;
  await metamask.openSidebarMenu(SidebarMenus.Advanced);

  await metamask.toggleDismissSecretRecoveryPhraseReminder();

  await metamask.addNetwork(LINEA_SEPOLIA_NETWORK);
  await metamask.switchNetwork("Sepolia", true);

  await metamask.page.click(metamask.homePage.selectors.accountMenu.accountButton);
  await metamask.page.locator(metamask.homePage.selectors.accountMenu.accountNames).last().click();
});
