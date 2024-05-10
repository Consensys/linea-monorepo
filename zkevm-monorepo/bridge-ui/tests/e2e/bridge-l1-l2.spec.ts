import { setupMetaMask, DappeteerPage, Dappeteer, DappeteerBrowser, launch } from '@chainsafe/dappeteer';
import config from '../config';
import { expect } from 'chai';
import { formatEther, formatUnits } from 'viem';
import { connectMetaMask, delay } from '../helpers/utils';
import { selectToken } from '../helpers/tokens';
import {
  getBridgeTransactionsCount,
  sendTokens,
  waitForTransactionListUpdate,
  waitForTransactionToConfirm,
} from '../helpers/transactions';

const TEST_URL = 'http://localhost:3000/';
const SEPOLIA_NETWORK = 'Sepolia';
const LINEA_SEPOLIA_NETWORK = 'Linea Sepolia';
const WEI_AMOUNT = formatEther(BigInt(1)).toString();
const USDC_AMOUNT = formatUnits(BigInt(1), 6).toString();
const TEST_TIMEOUT = 90_000;

let bridgePage: DappeteerPage<unknown>, metaMask: Dappeteer, browser: DappeteerBrowser<unknown, unknown>;

describe('Bridge L1 > L2', () => {
  test(
    'should set up the UI and metamask correctly',
    async () => {
      browser = await launch({
        automation: 'puppeteer',
        puppeteerOptions: { args: ['--no-sandbox'] },
        headless: true,
      });
      // We have to wait that the extension page is completely loaded
      await delay(1000);
      metaMask = await setupMetaMask(browser, {
        password: 'xxxxxxxxxxxxx',
      });
      //Opening new tab and opening the bridge UI web page
      bridgePage = await browser.newPage();
      await bridgePage.goto(TEST_URL, {
        waitUntil: 'networkidle',
      });
      await delay(1000);
      //Click on the native bridge button
      const nativeBridgeBtn = await bridgePage.waitForXPath("//button[@id='native-bridge-btn']");
      await nativeBridgeBtn.click();
      await delay(2000);

      //Agree to Terms
      const agreeBtn = await bridgePage.waitForXPath("//button[@id='agree-terms-btn']");
      await agreeBtn.click();
      await delay(1000);

      //Setting up account on MetaMask
      await metaMask.importPK(config.PRIVATE_KEY);

      //Opening Bridge UI tab
      await bridgePage.bringToFront();
      await delay(1000);

      //Connecting to MetaMask on Bridge UI
      await connectMetaMask(bridgePage, metaMask);
      await delay(1000);

      await bridgePage.bringToFront();
      await delay(1000);

      await metaMask.switchNetwork(SEPOLIA_NETWORK);

      await bridgePage.bringToFront();
    },
    TEST_TIMEOUT
  );

  test('should successfully go to the bridge UI page', async () => {
    const pageUrl = await bridgePage.url();
    expect(pageUrl, 'page url not correct').to.eq(`${TEST_URL}`);
  });

  test('should successfully display the correct heading', async () => {
    const header = 'Bridge';
    await bridgePage.waitForXPath("//h2[contains(text(),'" + header + "')]");
  });

  test('metamask should be connected to the right network', async () => {
    await bridgePage.waitForXPath("//span[@id='active-chain-name'][contains(text(),'" + SEPOLIA_NETWORK + "')]");
  });

  // TODO: enable this test once Sepolia has been added to the Bridge UI
  test.skip(
    'should be able to reload the transaction history',
    async () => {
      await bridgePage.bringToFront();
      await delay(1000);
      const reloadBtn = await bridgePage.waitForXPath("//button[@id='reload-history-btn']");
      await reloadBtn.click();
      await delay(1000);
      const reloadConfirmBtn = await bridgePage.waitForXPath("//button[@id='reload-history-confirm-btn']");
      await reloadConfirmBtn.click();
      await delay(1000);
      // This allows us to know when the reaload is completed
      await bridgePage.waitForXPath("//div[@id='transactions-list']//ul", {
        timeout: 70000,
      });
      await delay(1000);
    },
    TEST_TIMEOUT
  );

  test(
    'should be able to switch network',
    async () => {
      const chainSelectBtn = await bridgePage.waitForXPath("//summary[@id='chain-select']");
      await chainSelectBtn.click();
      await delay(1000);
      const alternativeChainBtn = await bridgePage.waitForXPath("//button[@id='switch-alternative-chain-btn']");
      await alternativeChainBtn.click();
      await delay(1000);
      await metaMask.switchNetwork(LINEA_SEPOLIA_NETWORK);
      await delay(2000);
      await bridgePage.bringToFront();
      await delay(1000);
    },
    TEST_TIMEOUT
  );
  
  // TODO: enable this test once Sepolia has been added to the Bridge UI
  test.skip(
    'should be able to claim funds if available',
    async () => {
      // Check if there is a claim button available
      const checkClaimBtn = await bridgePage.$$('#claim-funds-btn');
      if (checkClaimBtn.length > 0) {
        const claimBtn = await bridgePage.waitForXPath("//button[@id='claim-funds-btn']");
        await claimBtn.click();
        await delay(4000);
        await metaMask.confirmTransaction();
        await delay(1000);
        await waitForTransactionToConfirm(metaMask);
        await delay(1000);
      } else {
        console.warn('Claim funds could not be tested since no funds are waiting to be claimed');
      }
    },
    TEST_TIMEOUT
  );

  // TODO: enable this test once Sepolia has been added to the Bridge UI
  test.skip(
    'should be able to bridge ETH from L1 to L2',
    async () => {
      //Switch back to sepolia
      await metaMask.switchNetwork(SEPOLIA_NETWORK);
      await delay(1000);
      await bridgePage.bringToFront();
      const txnsLengthBefore = await getBridgeTransactionsCount(bridgePage);

      await sendTokens(bridgePage, WEI_AMOUNT, metaMask, true);
      await metaMask.confirmTransaction();

      // Wait for transaction to finish
      await waitForTransactionToConfirm(metaMask);

      await bridgePage.bringToFront();
      await delay(4000);
      // We check at the end that the transacton list is updated on the bridge UI
      const listUpdated = await waitForTransactionListUpdate(bridgePage, txnsLengthBefore);

      // Check that that new transaction is added to the list on the UI
      expect(listUpdated).to.eq(true);
    },
    TEST_TIMEOUT
  );
  
  // TODO: enable this test once Sepolia has been added to the Bridge UI
  test.skip(
    'should be able to bridge USDC from L1 to L2',
    async () => {
      const txnsLengthBefore = await getBridgeTransactionsCount(bridgePage);

      // Select USDC in the token list
      await selectToken(bridgePage, 'USDC');
      await delay(1000);

      await sendTokens(bridgePage, USDC_AMOUNT, metaMask);
      await delay(1000);

      await metaMask.confirmTransaction();
      await delay(1000);

      // Wait for transaction to finish
      await waitForTransactionToConfirm(metaMask);

      await bridgePage.bringToFront();
      // We check at the end that the transacton list is updated on the bridge UI
      const listUpdated = await waitForTransactionListUpdate(bridgePage, txnsLengthBefore);

      // Check that that new transaction is added to the list on the UI
      expect(listUpdated).to.eq(true);
    },
    TEST_TIMEOUT
  );

  // TODO: enable this test once Sepolia has been added to the Bridge UI
  test.skip(
    'should be able to bridge ERC20 tokens from L1 to L2',
    async () => {
      const txnsLengthBefore = await getBridgeTransactionsCount(bridgePage);

      // Select WETH in the token list (Easiest to get)
      await selectToken(bridgePage, 'WETH');

      await sendTokens(bridgePage, WEI_AMOUNT, metaMask);
      await delay(1000);
      await metaMask.confirmTransaction();
      await delay(1000);

      // Wait for transaction to finish
      await waitForTransactionToConfirm(metaMask);

      await bridgePage.bringToFront();
      // We check at the end that the transacton list is updated on the bridge UI
      const listUpdated = await waitForTransactionListUpdate(bridgePage, txnsLengthBefore);

      // Check that that new transaction is added to the list on the UI
      expect(listUpdated).to.eq(true);
    },
    TEST_TIMEOUT
  );
});
