import { Dappeteer, DappeteerPage } from '@chainsafe/dappeteer';
import { delay } from './utils';

const sendTokens = async (bridgePage: DappeteerPage<unknown>, amount: string, metaMask: Dappeteer, eth = false) => {
  const amountInput = await bridgePage.waitForXPath("//input[@id='amount-input']");

  // Sending the smallest amount of USDC
  await amountInput.type(amount);

  await delay(1000);

  // Check if there are funds available
  const approveBtnDisabled = await bridgePage.$$('#approve-btn.btn-disabled');
  const submitBtnDisabled = await bridgePage.$$('#submit-erc-btn.btn-disabled');
  if (approveBtnDisabled.length === 1 && submitBtnDisabled.length === 1) {
    throw 'No funds available, please add some funds before running the test';
  }

  const tokenType = eth ? 'eth' : 'erc';

  // Check that this amount has been approved
  if (tokenType === 'erc' && submitBtnDisabled.length === 1) {
    //We need to approve the amount first
    const approveBtn = await bridgePage.waitForXPath(`//button[@id='approve-btn']`);
    await delay(1000);
    await approveBtn.click();
    await metaMask.page.bringToFront();
    await delay(1000);
    await metaMask.page.reload();
    const nextBtn = await metaMask.page.waitForXPath("//button[contains(text(),'Next')]");
    nextBtn.click();
    await delay(2000);
    const approveMMBtn = await metaMask.page.waitForXPath("//button[contains(text(),'Approve')]");
    await approveMMBtn.click();
    await delay(1000);
    await waitForTransactionToConfirm(metaMask);
    await bridgePage.bringToFront();
    await delay(1000);
  }
  const submitBtn = await bridgePage.waitForXPath(`//button[@id='submit-${tokenType}-btn']`);
  await delay(1000);
  await submitBtn.click();
};

const waitForTransactionToConfirm = async (metaMask: Dappeteer) => {
  await metaMask.page.bringToFront();
  await metaMask.page.reload();
  const activityButton = await metaMask.page.waitForXPath("//button[contains(text(),'Activity')]");
  await activityButton.click();
  await delay(1000);

  let txCount = await metaMask.page.$$('.transaction-list__pending-transactions>.transaction-list-item');
  while (txCount.length > 0) {
    await delay(2000);
    txCount = await metaMask.page.$$('.transaction-list__pending-transactions>.transaction-list-item');
  }
};

const waitForTransactionListUpdate = async (bridgePage: DappeteerPage<unknown>, txCountBeforeUpdate: number) => {
  const maxTries = 20;
  let tryCount = 0;
  let listUpdated = false;
  do {
    await delay(2000);
    const newTxCount = await getBridgeTransactionsCount(bridgePage);

    listUpdated = newTxCount !== txCountBeforeUpdate;
    tryCount++;
  } while (!listUpdated && tryCount < maxTries);

  return listUpdated;
};

const getBridgeTransactionsCount = async (bridgePage: DappeteerPage<unknown>) => {
  const txns = await bridgePage.$$('#transactions-list>ul');
  return txns.length;
};

export { sendTokens, waitForTransactionToConfirm, getBridgeTransactionsCount, waitForTransactionListUpdate };
