import { Dappeteer, DappeteerPage } from '@chainsafe/dappeteer';

const delay = (time: number) => {
  return new Promise(function (resolve) {
    setTimeout(resolve, time);
  });
};

const connectMetaMask = async (bridgePage: DappeteerPage<unknown>, metaMask: Dappeteer) => {
  const connectMetamaskButton = await bridgePage.waitForXPath("//button[contains(text(),'Connect')]");
  await connectMetamaskButton.click();
  await delay(1000);
  const metaMaskButton = await bridgePage.$('pierce/wui-list-wallet:nth-child(3)');
  metaMaskButton?.click();
  await delay(2000);
  await metaMask.approve();
};

const requestAccounts = () => {
  window.ethereum.request<string[]>({ method: 'eth_requestAccounts' });
};

export { delay, connectMetaMask, requestAccounts };
