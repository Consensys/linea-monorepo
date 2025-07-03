import { LOCAL_L1_NETWORK, LOCAL_L2_NETWORK, METAMASK_PASSWORD, METAMASK_SEED_PHRASE } from "../constants";
import { defineWalletSetup } from "@synthetixio/synpress";
import { MetaMask, getExtensionId } from "@synthetixio/synpress/playwright";

export default defineWalletSetup(METAMASK_PASSWORD, async (context, walletPage) => {
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  //@ts-ignore
  const extensionId = await getExtensionId(context, "MetaMask");
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  //@ts-ignore
  const metamask = new MetaMask(context, walletPage, METAMASK_PASSWORD, extensionId);
  await metamask.importWallet(METAMASK_SEED_PHRASE);

  await metamask.openSettings();

  const SidebarMenus = metamask.homePage.selectors.settings.SettingsSidebarMenus;
  await metamask.openSidebarMenu(SidebarMenus.Advanced);

  await metamask.toggleDismissSecretRecoveryPhraseReminder();

  await metamask.addNetwork(LOCAL_L1_NETWORK);
  await metamask.addNetwork(LOCAL_L2_NETWORK);

  await metamask.switchNetwork("L1", true);

  // We need to use the Account 7 to not conflict with accounts used by the local docker stack
  for (let i = 2; i < 8; i++) {
    await metamask.addNewAccount(`Account ${i}`);
  }

  await metamask.switchAccount("Account 7");
});
