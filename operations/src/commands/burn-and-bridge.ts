import { Command, Flags } from "@oclif/core";
import {
  createPublicClient,
  formatEther,
  formatUnits,
  Hex,
  http,
  parseEventLogs,
  parseSignature,
  serializeTransaction,
  TransactionSerializable,
} from "viem";
import { linea } from "viem/chains";
import { Agent } from "https";
import { fromZonedTime } from "date-fns-tz";
import { addSeconds } from "date-fns";
import { estimateTransactionGas, sendTransaction } from "../utils/common/transactions.js";
import { getWeb3SignerSignature } from "../utils/common/signature.js";
import { buildHttpsAgent } from "../utils/common/https-agent.js";
import { validateUrl } from "../utils/common/validation.js";
import { address, hexString } from "../utils/common/custom-flags.js";
import {
  computeBurnAndBridgeCalldata,
  computeSwapCalldata,
  getInvoiceArrears,
  getMinimumFee,
  getQuote,
} from "../utils/burn-and-bridge/contract.js";
import { LINEA_TOKEN_ADDRESS, WETH_TOKEN_ADDRESS } from "../utils/burn-and-bridge/constants.js";
import { ETH_BURNT_SWAPPED_AND_BRIDGED_EVENT_ABI } from "../utils/burn-and-bridge/abi.js";

export default class BurnAndBridge extends Command {
  static examples = [
    `<%= config.bin %> <%= command.id %> 
      --senderAddress=0xYourSenderAddress
      --rollupRevenueVaultContractAddress=0xYourContractAddress
      --quoteContractAddress=0xYourQuoteContractAddress
      --rpcUrl=https://mainnet.infura.io/v3/YOUR-PROJECT-ID
      --web3SignerUrl=http://localhost:8545
      --web3SignerPublicKey=0xYourWeb3SignerPublicKey
      --tls
      --web3SignerKeystorePath=/path/to/keystore.p12
      --web3SignerKeystorePassphrase=yourPassphrase
      --web3SignerTrustedStorePath=/path/to/ca.p12
      --web3SignerTrustedStorePassphrase=yourTrustedStorePassphrase
      --swapAmountSlippageBps=50
    `,
    // Dry run
    `<%= config.bin %> <%= command.id %>
      --senderAddress=0xYourSenderAddress
      --rollupRevenueVaultContractAddress=0xYourContractAddress
      --quoteContractAddress=0xYourQuoteContractAddress
      --rpcUrl=https://mainnet.infura.io/v3/YOUR-PROJECT-ID
      --web3SignerUrl=http://localhost:8545
      --web3SignerPublicKey=0xYourWeb3SignerPublicKey
      --tls
      --web3SignerKeystorePath=/path/to/keystore.p12
      --web3SignerKeystorePassphrase=yourPassphrase
      --web3SignerTrustedStorePath=/path/to/ca.p12
      --web3SignerTrustedStorePassphrase=yourTrustedStorePassphrase
      --swapAmountSlippageBps=50
      --dryRun
    `,
  ];

  static strict = true;

  static flags = {
    senderAddress: address({
      char: "s",
      description: "Sender address",
      required: true,
      env: "BURN_AND_BRIDGE_SENDER_ADDRESS",
    }),
    rollupRevenueVaultContractAddress: address({
      description: "Rollup Revenue Vault Contract address",
      required: true,
      env: "BURN_AND_BRIDGE_ROLLUP_REVENUE_VAULT_CONTRACT_ADDRESS",
    }),
    l2MessageServiceContractAddress: address({
      description: "L2 Message Service Contract address",
      required: true,
      env: "BURN_AND_BRIDGE_L2_MESSAGE_SERVICE_CONTRACT_ADDRESS",
    }),
    quoteContractAddress: address({
      description: "Quote Contract address",
      required: true,
      env: "BURN_AND_BRIDGE_QUOTE_CONTRACT_ADDRESS",
    }),
    rpcUrl: Flags.string({
      description: "Blockchain RPC URL",
      required: true,
      parse: async (input) => validateUrl("blockchain-rpc-url", input, ["http:", "https:"]),
      env: "BURN_AND_BRIDGE_BLOCKCHAIN_RPC_URL",
    }),
    web3SignerUrl: Flags.string({
      description: "Web3 Signer URL",
      required: true,
      parse: async (input) => validateUrl("web3-signer-url", input, ["http:", "https:"]),
      env: "BURN_AND_BRIDGE_WEB3_SIGNER_URL",
    }),
    web3SignerPublicKey: hexString({
      description: "Web3 Signer Public Key",
      required: true,
      env: "BURN_AND_BRIDGE_WEB3_SIGNER_PUBLIC_KEY",
    }),
    dryRun: Flags.boolean({
      description: "Dry run flag",
      required: false,
      default: false,
      env: "BURN_AND_BRIDGE_DRY_RUN",
    }),
    tls: Flags.boolean({
      description: "Enable TLS",
      required: false,
      default: false,
      env: "BURN_AND_BRIDGE_TLS",
      relationships: [
        {
          type: "all",
          flags: [
            { name: "web3SignerKeystorePath", when: async (flags) => flags["tls"] === true },
            { name: "web3SignerKeystorePassphrase", when: async (flags) => flags["tls"] === true },
            { name: "web3SignerTrustedStorePath", when: async (flags) => flags["tls"] === true },
            { name: "web3SignerTrustedStorePassphrase", when: async (flags) => flags["tls"] === true },
          ],
        },
      ],
    }),
    web3SignerKeystorePath: Flags.string({
      description: "Path to the web3 signer keystore file",
      required: false,
      env: "BURN_AND_BRIDGE_WEB3_SIGNER_KEYSTORE_PATH",
    }),
    web3SignerKeystorePassphrase: Flags.string({
      description: "Passphrase for the web3 signer keystore",
      required: false,
      env: "BURN_AND_BRIDGE_WEB3_SIGNER_KEYSTORE_PASSPHRASE",
    }),
    web3SignerTrustedStorePath: Flags.string({
      description: "Path to the web3 signer trusted store file",
      required: false,
      env: "BURN_AND_BRIDGE_WEB3_SIGNER_TRUSTED_STORE_PATH",
    }),
    web3SignerTrustedStorePassphrase: Flags.string({
      description: "Passphrase for the web3 signer trusted store file",
      required: false,
      env: "BURN_AND_BRIDGE_WEB3_SIGNER_TRUSTED_STORE_PASSPHRASE",
    }),
    swapAmountSlippageBps: Flags.integer({
      description: "Allowed slippage in basis points for the swap",
      required: true,
      env: "BURN_AND_BRIDGE_SWAP_AMOUNT_SLIPPAGE_BPS",
    }),
  };

  public async run(): Promise<void> {
    const { flags } = await this.parse(BurnAndBridge);
    const {
      senderAddress,
      rpcUrl,
      rollupRevenueVaultContractAddress,
      l2MessageServiceContractAddress,
      quoteContractAddress,
      web3SignerUrl,
      web3SignerPublicKey,
      web3SignerKeystorePath,
      web3SignerKeystorePassphrase,
      web3SignerTrustedStorePath,
      web3SignerTrustedStorePassphrase,
      tls,
      dryRun,
    } = flags;

    const client = createPublicClient({
      chain: linea,
      transport: http(rpcUrl, { batch: true, retryCount: 3 }),
    });

    const invoiceArrearsResult = await getInvoiceArrears(client, rollupRevenueVaultContractAddress);

    const invoiceArrearsInWei = invoiceArrearsResult.match(
      (value) => value,
      (error) => this.error(`Failed to get invoice arrears: ${error.message}`),
    );

    if (invoiceArrearsInWei === 0n) {
      this.log(`No invoice arrears to pay. Exiting.`);
      return;
    }

    const rollupRevenueVaultContractBalance = await client.getBalance({
      address: rollupRevenueVaultContractAddress,
    });

    const minimumFeeResult = await getMinimumFee(client, l2MessageServiceContractAddress);

    const minimumFeeInWei = minimumFeeResult.match(
      (value) => value,
      (error) => this.error(`Failed to get minimum fee: ${error.message}`),
    );

    if (rollupRevenueVaultContractBalance < minimumFeeInWei) {
      this.error(
        `Insufficient balance in Rollup Revenue Vault Contract to cover the minimum fee. balance=${formatEther(rollupRevenueVaultContractBalance)} ETH minimumFee=${formatEther(minimumFeeInWei)} ETH`,
      );
    }

    const balanceAvailable = rollupRevenueVaultContractBalance - minimumFeeInWei;
    const ETH_BURNT_PERCENTAGE = 20n; // 20% of the ETH balance will be burnt
    const ethToBurn = (balanceAvailable * ETH_BURNT_PERCENTAGE) / 100n;

    const amountToBeSwappedInWei = balanceAvailable - ethToBurn;

    const quoteResult = await getQuote(client, quoteContractAddress, {
      tokenIn: WETH_TOKEN_ADDRESS,
      tokenOut: LINEA_TOKEN_ADDRESS,
      amountIn: amountToBeSwappedInWei,
      tickSpacing: 50,
      sqrtPriceLimitX96: 0n,
    });

    const [minAmountOut] = quoteResult.match(
      (value) => value,
      (error) => this.error(`Failed to get quote: ${error.message}`),
    );

    const slippageBps = BigInt(flags.swapAmountSlippageBps);
    const minLineaOut = (minAmountOut * (10_000n - slippageBps)) / 10_000n;

    this.log(
      `Minimum LINEA out (after slippage): minLineaOut=${formatUnits(minLineaOut, 18)} LINEA slippageBps=${slippageBps}`,
    );

    /******************************
      TRANSACTION GAS ESTIMATION
     ******************************/

    const deadline = addSeconds(fromZonedTime(Math.floor(new Date().getTime()), "UTC").getTime() / 1000, 3);
    const swapCalldata = computeSwapCalldata(minLineaOut, BigInt(Math.floor(deadline.getTime() / 1000)));
    const burnAndBridgeCalldata = computeBurnAndBridgeCalldata(swapCalldata);

    const gasEstimationResult = await estimateTransactionGas(client, {
      to: rollupRevenueVaultContractAddress,
      account: senderAddress,
      value: 0n,
      data: burnAndBridgeCalldata,
    });

    const { gasLimit, baseFeePerGas, priorityFeePerGas } = gasEstimationResult.match(
      (value) => value,
      (error) => this.error(`Failed to estimate gas: ${error.message}`),
    );

    this.log(
      `Gas estimation for burnAndBridge transaction: gasLimit=${gasLimit} baseFeePerGas=${baseFeePerGas} priorityFeePerGas=${priorityFeePerGas}`,
    );

    /******************************
          SIGNING TRANSACTION
     ******************************/

    let httpsAgent: Agent | undefined;
    if (
      tls &&
      web3SignerKeystorePath &&
      web3SignerKeystorePassphrase &&
      web3SignerTrustedStorePath &&
      web3SignerTrustedStorePassphrase
    ) {
      this.log(`Using TLS for secure communication with Web3 Signer.`);
      httpsAgent = buildHttpsAgent(
        web3SignerKeystorePath,
        web3SignerKeystorePassphrase,
        web3SignerTrustedStorePath,
        web3SignerTrustedStorePassphrase,
      );
    }

    const transactionToSerialize: TransactionSerializable = {
      to: rollupRevenueVaultContractAddress,
      type: "eip1559",
      value: 0n,
      data: burnAndBridgeCalldata,
      chainId: linea.id,
      gas: gasLimit,
      maxFeePerGas: baseFeePerGas + priorityFeePerGas,
      maxPriorityFeePerGas: priorityFeePerGas,
    };

    const signature = await this.signBurnAndBridgeTransaction(
      web3SignerUrl,
      web3SignerPublicKey,
      transactionToSerialize,
      httpsAgent,
    );

    if (dryRun) {
      this.log(`Dry run mode - transaction not submitted.`);
      return;
    }

    /******************************
        BROADCASTING TRANSACTION
     ******************************/

    const serializeSignedTransaction = serializeTransaction(transactionToSerialize, parseSignature(signature));

    this.log(`Broadcasting submitInvoice transaction to the network...`);
    const transactionResult = await sendTransaction(client, serializeSignedTransaction);
    const receipt = transactionResult.match(
      (value) => value,
      (error) => this.error(`Failed to send transaction: ${error.message}`),
    );

    if (receipt.status === "reverted") {
      this.error(`Burn and bridge failed. transactionHash=${receipt.transactionHash}`);
    }

    const [event] = parseEventLogs({
      abi: ETH_BURNT_SWAPPED_AND_BRIDGED_EVENT_ABI,
      logs: receipt.logs,
      eventName: "EthBurntSwappedAndBridged",
    });

    this.log(
      `Burn and bridge transaction successfully processed. transactionHash=${receipt.transactionHash} ethBurnt=${formatEther(event.args.ethBurnt)} lineaTokensBridged=${formatUnits(event.args.lineaTokensBridged, 18)}`,
    );
  }

  /**
   * Sign the burn and bridge transaction using Web3 Signer.
   * @param web3SignerUrl Web3 Signer URL.
   * @param web3SignerPublicKey Web3 Signer Public Key.
   * @param transactionToSerialize Transaction to be serialized and signed.
   * @param httpsAgent Optional HTTPS Agent for secure communication.
   * @returns The signature as a hex string.
   */
  public async signBurnAndBridgeTransaction(
    web3SignerUrl: string,
    web3SignerPublicKey: string,
    transactionToSerialize: TransactionSerializable,
    httpsAgent?: Agent,
  ): Promise<Hex> {
    const signatureResult = await getWeb3SignerSignature(
      web3SignerUrl,
      web3SignerPublicKey,
      transactionToSerialize,
      httpsAgent,
    );

    return signatureResult.match(
      (value) => value,
      (error) => this.error(`Failed to get signature from Web3 Signer: ${error.message}`),
    );
  }
}
