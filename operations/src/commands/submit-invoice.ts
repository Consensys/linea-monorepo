import { Command, Flags } from "@oclif/core";
import {
  Address,
  createPublicClient,
  Hex,
  http,
  parseEther,
  parseEventLogs,
  parseSignature,
  serializeTransaction,
  TransactionSerializable,
} from "viem";
import { linea } from "viem/chains";
import { Agent } from "https";
import { GetCostAndUsageCommandInput } from "@aws-sdk/client-cost-explorer";
import { formatInTimeZone } from "date-fns-tz";
import { addDays } from "date-fns";
import { computeInvoicePeriod } from "../utils/eth-transfer/time.js";
import { generateQueryParameters, getDuneClient, runDuneQuery } from "../utils/common/dune.js";
import { estimateTransactionGas, sendTransaction } from "../utils/eth-transfer/transactions.js";
import { getWeb3SignerSignature } from "../utils/common/signature.js";
import { INVOICE_PROCESSED_EVENT_ABI } from "../utils/eth-transfer/constants.js";
import { getHttpsAgent } from "../utils/common/https-agent.js";
import { createAwsCostExplorerClient, getDailyAwsCosts } from "../utils/common/aws.js";
import { computeSubmitInvoiceCalldata, getLastInvoiceDate } from "../utils/eth-transfer/contract.js";
import { validateEthereumAddress, validateHexString, validateUrl } from "../utils/common/validation.js";

export const address = Flags.custom<Address>({
  parse: async (input) => validateEthereumAddress("address", input),
});

export const hexString = Flags.custom<Hex>({
  parse: async (input) => validateHexString("hex-string", input),
});

export const awsCostsApiFilters = Flags.custom<GetCostAndUsageCommandInput>({
  parse: (input: string) => {
    try {
      return JSON.parse(input);
    } catch {
      throw new Error("Invalid JSON string for AWS Costs API Filters");
    }
  },
});

export default class SubmitInvoice extends Command {
  static examples = [
    `<%= config.bin %> <%= command.id %> 
      --senderAddress=0xYourSenderAddress
      --contractAddress=0xYourContractAddress
      --periodDays=10
      --reportingLagDays=2
      --rpcUrl=https://mainnet.infura.io/v3/YOUR-PROJECT-ID
      --web3SignerUrl=http://localhost:8545
      --web3SignerPublicKey=0xYourWeb3SignerPublicKey
      --duneApiKey=YOUR_DUNE_KEY
      --duneQueryId=12345
      --tls
      --web3SignerKeystorePath=/path/to/keystore.p12
      --web3SignerKeystorePassphrase=yourPassphrase
      --web3SignerTrustedStorePath=/path/to/ca.p12
      --web3SignerTrustedStorePassphrase=yourTrustedStorePassphrase
      --awsCostsApiFilters='{"Granularity":"DAILY","Metrics":["AmortizedCost"],"GroupBy":[]}'
    `,
    // Dry run
    `<%= config.bin %> <%= command.id %>
      --senderAddress=0xYourSenderAddress
      --contractAddress=0xYourContractAddress
      --periodDays=10
      --reportingLagDays=2
      --rpcUrl=https://mainnet.infura.io/v3/YOUR-PROJECT-ID
      --web3SignerUrl=http://localhost:8545
      --web3SignerPublicKey=0xYourWeb3SignerPublicKey
      --duneApiKey=YOUR_DUNE_KEY
      --duneQueryId=12345
      --tls
      --web3SignerKeystorePath=/path/to/keystore.p12
      --web3SignerKeystorePassphrase=yourPassphrase
      --web3SignerTrustedStorePath=/path/to/ca.p12
      --web3SignerTrustedStorePassphrase=yourTrustedStorePassphrase
      --awsCostsApiFilters='{"Granularity":"DAILY","Metrics":["AmortizedCost"],"GroupBy":[]}'
      --dryRun
    `,
  ];

  static strict = true;

  static flags = {
    senderAddress: address({
      char: "s",
      description: "Sender address",
      required: true,
      env: "ETH_TRANSFER_SENDER_ADDRESS",
    }),
    contractAddress: address({
      char: "c",
      description: "Contract address",
      required: true,
      env: "ETH_TRANSFER_CONTRACT_ADDRESS",
    }),
    periodDays: Flags.integer({
      char: "p",
      description: "Period in days for invoice generation",
      required: true,
      parse: async (input) => parseInt(input),
      env: "ETH_TRANSFER_PERIOD_DAYS",
    }),
    reportingLagDays: Flags.integer({
      char: "r",
      description: "Reporting lag in days for invoice generation",
      required: true,
      parse: async (input) => parseInt(input),
      env: "ETH_TRANSFER_REPORTING_LAG_DAYS",
    }),
    rpcUrl: Flags.string({
      description: "Blockchain RPC URL",
      required: true,
      parse: async (input) => validateUrl("blockchain-rpc-url", input, ["http:", "https:"]),
      env: "ETH_TRANSFER_BLOCKCHAIN_RPC_URL",
    }),
    web3SignerUrl: Flags.string({
      description: "Web3 Signer URL",
      required: true,
      parse: async (input) => validateUrl("web3-signer-url", input, ["http:", "https:"]),
      env: "ETH_TRANSFER_WEB3_SIGNER_URL",
    }),
    web3SignerPublicKey: hexString({
      description: "Web3 Signer Public Key",
      required: true,
      env: "ETH_TRANSFER_WEB3_SIGNER_PUBLIC_KEY",
    }),
    dryRun: Flags.boolean({
      description: "Dry run flag",
      required: false,
      default: false,
      env: "ETH_TRANSFER_DRY_RUN",
    }),
    tls: Flags.boolean({
      description: "Enable TLS",
      required: false,
      default: false,
      env: "ETH_TRANSFER_TLS",
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
      env: "ETH_TRANSFER_WEB3_SIGNER_KEYSTORE_PATH",
    }),
    web3SignerKeystorePassphrase: Flags.string({
      description: "Passphrase for the web3 signer keystore",
      required: false,
      env: "ETH_TRANSFER_WEB3_SIGNER_KEYSTORE_PASSPHRASE",
    }),
    web3SignerTrustedStorePath: Flags.string({
      description: "Path to the web3 signer trusted store file",
      required: false,
      env: "ETH_TRANSFER_WEB3_SIGNER_TRUSTED_STORE_PATH",
    }),
    web3SignerTrustedStorePassphrase: Flags.string({
      description: "Passphrase for the web3 signer trusted store file",
      required: false,
      env: "ETH_TRANSFER_WEB3_SIGNER_TRUSTED_STORE_PASSPHRASE",
    }),
    duneApiKey: Flags.string({
      description: "Dune API Key",
      required: true,
      env: "ETH_TRANSFER_DUNE_API_KEY",
    }),
    duneQueryId: Flags.integer({
      description: "Dune Query ID",
      required: true,
      parse: async (input) => parseInt(input),
      env: "ETH_TRANSFER_DUNE_QUERY_ID",
    }),
    awsCostsApiFilters: awsCostsApiFilters({
      description: "AWS Costs API Filters as JSON string",
      required: true,
      env: "ETH_TRANSFER_AWS_COSTS_API_FILTERS",
    }),
  };

  public async run(): Promise<void> {
    const { flags } = await this.parse(SubmitInvoice);
    const {
      senderAddress,
      rpcUrl,
      contractAddress,
      periodDays,
      reportingLagDays,
      web3SignerUrl,
      web3SignerPublicKey,
      web3SignerKeystorePath,
      web3SignerKeystorePassphrase,
      web3SignerTrustedStorePath,
      web3SignerTrustedStorePassphrase,
      tls,
      dryRun,
      awsCostsApiFilters,
    } = flags;

    const client = createPublicClient({
      chain: linea,
      transport: http(rpcUrl, { batch: true, retryCount: 3 }),
    });

    /******************************
      INVOICE PERIOD COMPUTATION
     ******************************/

    const lastInvoiceDate = await getLastInvoiceDate(client, contractAddress);

    if (lastInvoiceDate.isErr()) {
      this.error(`Failed to retrieve the last invoice date from the contract: ${lastInvoiceDate.error.message}`);
    }

    this.log(`Last invoice date (timestamp in seconds): ${lastInvoiceDate.value}`);

    const currentTimestampInSeconds = Math.floor(Date.now() / 1000);
    const invoicePeriod = computeInvoicePeriod(
      Number(lastInvoiceDate.value),
      currentTimestampInSeconds,
      periodDays,
      reportingLagDays,
    );

    if (!invoicePeriod) {
      this.warn("No invoice to process at this time.");
      return;
    }

    this.log(
      `Invoice period to process: from ${invoicePeriod.startDate.toISOString()} to ${invoicePeriod.endDate.toISOString()}`,
    );

    /******************************
            AWS COSTS FETCHING
     ******************************/

    const awsClient = createAwsCostExplorerClient({ region: "us-east-1" });
    this.log(
      `Fetching AWS costs for the invoice period from=${formatInTimeZone(invoicePeriod.startDate, "UTC", "yyyy-MM-dd")} to=${formatInTimeZone(addDays(invoicePeriod.endDate, 1), "UTC", "yyyy-MM-dd")}`,
    );

    const awsCostsResult = await getDailyAwsCosts(awsClient, {
      ...awsCostsApiFilters,
      TimePeriod: {
        Start: formatInTimeZone(invoicePeriod.startDate, "UTC", "yyyy-MM-dd"),
        End: formatInTimeZone(addDays(invoicePeriod.endDate, 1), "UTC", "yyyy-MM-dd"),
      },
    });

    const awsCostsInUsd = awsCostsResult.match(
      (value) => {
        const { ResultsByTime } = value;

        if (!ResultsByTime || ResultsByTime.length === 0) {
          this.error("No AWS cost data returned for the specified period.");
        }

        if (!awsCostsApiFilters.Metrics || awsCostsApiFilters.Metrics.length !== 1) {
          this.error("AWS Costs API Filters must specify one metric.");
        }

        const totalDailyCosts = ResultsByTime[0].Total?.[awsCostsApiFilters.Metrics[0]].Amount;

        if (!totalDailyCosts) {
          this.error("No AWS cost data found for the specified metric.");
        }

        return parseFloat(totalDailyCosts);
      },
      (error) => this.error(`Failed to fetch AWS costs: ${error.message}`),
    );

    this.log(`Total AWS costs in USD for the period: ${awsCostsInUsd}`);

    /******************************
        ONCHAIN COSTS FETCHING
     ******************************/

    const duneClient = getDuneClient(flags.duneApiKey);
    const onChainCostsResult = await runDuneQuery(
      duneClient,
      flags.duneQueryId,
      generateQueryParameters({
        fromDate: invoicePeriod.startDate,
        toDate: invoicePeriod.endDate,
      }),
    );

    const onChainCostsInEth = onChainCostsResult.match(
      (value) => {
        const { result } = value;
        if (!result || !result.rows || result.rows.length === 0) {
          this.error("No Dune query result returned for the specified period.");
        }
        return result.rows.reduce((acc, row) => acc + (row.total_costs_per_day as number), 0);
      },
      (error) => this.error(`Failed to run Dune query: ${error.message}`),
    );

    this.log(`Total on-chain costs in ETH for the period: ${onChainCostsInEth}`);

    /******************************
        TOTAL COSTS COMPUTATION
     ******************************/

    // TODO: convert awsCostsInUsd to ETH using some oracle or API
    const awsCostsInEth = awsCostsInUsd;
    const totalCostsInEth = parseEther((awsCostsInEth + onChainCostsInEth).toString());

    if (totalCostsInEth === 0n) {
      this.warn("No costs to process at this time.");
      this.warn("Please check your Dune query and AWS costs API calls.");
      return;
    }

    this.log(`Total costs to invoice in ETH: ${totalCostsInEth}`);

    /******************************
      TRANSACTION GAS ESTIMATION
     ******************************/

    const submitInvoiceCalldata = computeSubmitInvoiceCalldata(
      BigInt(Math.floor(invoicePeriod.startDate.getTime() / 1000)),
      BigInt(Math.floor(invoicePeriod.endDate.getTime() / 1000)),
      totalCostsInEth,
    );

    const gasEstimationResult = await estimateTransactionGas(client, {
      to: contractAddress,
      account: senderAddress,
      value: 0n,
      data: submitInvoiceCalldata,
    });

    const { gasLimit, baseFeePerGas, priorityFeePerGas } = gasEstimationResult.match(
      (value) => value,
      (error) => this.error(`Failed to estimate gas: ${error.message}`),
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
      httpsAgent = getHttpsAgent(
        web3SignerKeystorePath,
        web3SignerKeystorePassphrase,
        web3SignerTrustedStorePath,
        web3SignerTrustedStorePassphrase,
      );
    }

    const transactionToSerialize: TransactionSerializable = {
      to: contractAddress,
      type: "eip1559",
      value: BigInt(totalCostsInEth),
      data: submitInvoiceCalldata,
      chainId: linea.id,
      gas: gasLimit,
      maxFeePerGas: baseFeePerGas + priorityFeePerGas,
      maxPriorityFeePerGas: priorityFeePerGas,
    };

    const signature = await getWeb3SignerSignature(
      web3SignerUrl,
      web3SignerPublicKey,
      transactionToSerialize,
      httpsAgent,
    );

    if (signature.isErr()) {
      this.error(`Failed to get signature from Web3 Signer: ${signature.error.message}`);
    }

    if (dryRun) {
      this.log(`Dry run mode - transaction not submitted.`);
      return;
    }

    /******************************
        BROADCASTING TRANSACTION
     ******************************/

    const serializeSignedTransaction = serializeTransaction(transactionToSerialize, parseSignature(signature.value));

    this.log(`Broadcasting submitInvoice transaction to the network...`);
    const transactionResult = await sendTransaction(client, serializeSignedTransaction);
    const receipt = transactionResult.match(
      (value) => value,
      (error) => this.error(`Failed to send transaction: ${error.message}`),
    );

    if (receipt.status === "success") {
      const [event] = parseEventLogs({
        abi: INVOICE_PROCESSED_EVENT_ABI,
        logs: receipt.logs,
        eventName: "InvoiceProcessed",
      });

      this.log(
        `Invoice successfully submitted: transactionHash=${receipt.transactionHash} eventName=${event.eventName} receiver=${event.args.receiver} startTimestamp=${event.args.startTimestamp} endTimestamp=${event.args.endTimestamp} amountPaid=${event.args.amountPaid} amountRequested=${event.args.amountRequested}`,
      );
      return;
    }

    this.error(`Invoice submission failed. transactionHash=${receipt.transactionHash}`);
  }
}
