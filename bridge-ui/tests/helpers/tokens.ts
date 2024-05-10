import { DappeteerPage } from '@chainsafe/dappeteer';
import { delay } from './utils';

const selectToken = async (bridgePage: DappeteerPage<unknown>, tokenName: string) => {
  const tokenETHBtn = await bridgePage.waitForXPath("//button[@id='token-select-btn']");
  await tokenETHBtn.click();
  await delay(1000);
  const tokenUSDCBtn = await bridgePage.waitForXPath(`//button[@id='token-details-${tokenName}-btn']`);
  await tokenUSDCBtn.click();
};

export { selectToken };
