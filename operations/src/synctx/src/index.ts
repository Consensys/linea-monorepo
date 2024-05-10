import { Command } from "@oclif/core";
import { BigNumber, UnsignedTransaction, ethers } from "ethers";
import { Flags } from "@oclif/core";
import { readFileSync } from "fs";

type Txpool = {
  pending: object;
  queued: object;
};

//TODO: add pagination for besu?
const clientApi: { [key: string]: { api: string; params: Array<any> } } = {
  geth: { api: "txpool_content", params: [] },
  besu: { api: "txpool_besuPendingTransactions", params: [2000] },
};

const isValidNodeTarget = (sourceNode: string, targetNode: string) => {
  try {
    /* eslint-disable no-new */
    if (sourceNode) {
      new URL(sourceNode);
    }
    new URL(targetNode);
    return true;
  } catch {}

  return false;
};

const isLocalPort = (value: string) => {
  try {
    const port = Number(value);
    return port >= 1 && port <= 65535;
  } catch {}

  return false;
};

const parsePendingTransactions = (pool: Txpool) => {
  return (Object.values(pool.pending).map((el: any) => Object.values(el)) as unknown as any[]).flat();
};

const getPendingTransactions = (sourcePool: any, targetPool: any) => {
  // calculate diff between source and target
  const targetPendingTransactions = new Set();

  targetPool.forEach((tx: any) => {
    targetPendingTransactions.add(tx.hash);
  });

  return sourcePool.filter((tx: any) => !targetPendingTransactions.has(tx.hash));
};

const getClientType = async (nodeProvider: ethers.providers.JsonRpcProvider) => {
  // Fetch client type of Besu or Geth
  const res: string = await nodeProvider.send("web3_clientVersion", []);

  const clientType = res.slice(0, 4).toLowerCase();
  if (!["geth", "besu"].includes(clientType)) {
    throw new Error(`Invalid node client type, must be either geth or besu`);
  }
  return clientType;
};

export default class Sync extends Command {
  static examples = [
    "<%= config.bin %> <%= command.id %> --source=8500 --target=8501 --local",
    "<%= config.bin %> <%= command.id %> --source=http://geth-archive-1:8545 --target=http://geth-validator-1:8545 --concurrency=10",
  ];

  static flags = {
    source: Flags.string({
      char: "s",
      description: "source node to sync from, mutually exclusive with file flag",
      helpGroup: "Node",
      multiple: false,
      required: true,
      default: "",
      env: "SYNCTX_SOURCE",
    }),
    target: Flags.string({
      char: "t",
      description: "target node to sync to",
      helpGroup: "Node",
      multiple: false,
      required: true,
      env: "SYNCTX_TARGET",
    }),
    "dry-run": Flags.boolean({
      description: "enable dry run",
      helpGroup: "Config",
      required: false,
      default: false,
      env: "SYNCTX_DRYRUN",
    }),
    local: Flags.boolean({
      description: "enable local run, provide only forwarded ports",
      helpGroup: "Config",
      required: false,
      default: false,
      env: "SYNCTX_LOCAL",
    }),
    concurrency: Flags.integer({
      char: "c",
      description: "number of concurrent batch requests",
      helpGroup: "Config",
      multiple: false,
      required: false,
      default: 10,
      env: "SYNCTX_CONCURRENCY",
    }),
    file: Flags.string({
      char: "f",
      description: "local txs file to read from, mutually exclusive with source flag",
      helpGroup: "Config",
      multiple: false,
      requiredOrDefaulted: true,
      default: "",
      env: "SYNCTX_FILE",
    }),
  };
  public async run(): Promise<void> {
    const { flags } = await this.parse(Sync);

    let sourceNode: string = flags.source;
    let targetNode: string = flags.target;
    const filePath: string = flags.file;
    let pendingTransactionsToSync: any[] = [];
    const concurrentCount = flags.concurrency as number;

    if ((filePath === "" && sourceNode === "") || (filePath !== "" && sourceNode !== "")) {
      this.error(
        "Invalid flag values are supplied, source and file are mutually exclusive and at least one needs to be specified",
      );
    }

    if (flags.local) {
      sourceNode = sourceNode && isLocalPort(sourceNode) ? `http://localhost:${sourceNode}` : sourceNode;
      targetNode = isLocalPort(targetNode) ? `http://localhost:${targetNode}` : targetNode;
    }

    if (!isValidNodeTarget(sourceNode, targetNode)) {
      this.error("Invalid nodes supplied to source and/or target, must be valid URL");
    }

    const sourceProvider = sourceNode ? new ethers.providers.JsonRpcProvider(sourceNode) : undefined;
    const targetProvider = new ethers.providers.JsonRpcProvider(targetNode);

    const sourceClientType = sourceProvider ? await getClientType(sourceProvider) : undefined;
    const targetClientType = await getClientType(targetProvider);

    if (sourceNode) {
      this.log(`Source ${sourceClientType} node: ${sourceNode}`);
    } else {
      this.log(`Skip checking source node type as txs file is supplied`);
    }
    this.log(`Target ${targetClientType} node: ${targetNode}`);

    let sourceTransactionPool: Txpool = { pending: {}, queued: {} };
    let targetTransactionPool: Txpool = { pending: {}, queued: {} };
    try {
      if (sourceProvider && sourceClientType) {
        this.log(`Fetching pending txs from txpool`);
        sourceTransactionPool = await sourceProvider.send(
          clientApi[sourceClientType].api,
          clientApi[sourceClientType].params,
        );
        targetTransactionPool = await targetProvider.send(
          clientApi[targetClientType].api,
          clientApi[targetClientType].params,
        );
      } else {
        this.log(`Skip fetching txs from source node as txs file is supplied`);
      }
    } catch (err) {
      this.error(`Invalid rpc provider(s) - ${err}`);
    }

    if (
      (sourceClientType === "geth" && Object.keys(sourceTransactionPool.pending).length === 0) ||
      (sourceClientType === "besu" && Object.keys(sourceTransactionPool).length === 0)
    ) {
      this.log("No pending transactions found on source node");
      return;
    }

    const sourcePendingTransactions: any =
      sourceClientType === "geth" ? parsePendingTransactions(sourceTransactionPool) : sourceTransactionPool;
    const targetPendingTransactions: any =
      targetClientType === "geth" ? parsePendingTransactions(targetTransactionPool) : targetTransactionPool;

    if (sourceNode) {
      this.log(`Source pending transactions: ${sourcePendingTransactions.length}`);
      this.log(`Target pending transactions: ${targetPendingTransactions.length}`);
    }

    pendingTransactionsToSync = sourceNode
      ? getPendingTransactions(sourcePendingTransactions, targetPendingTransactions)
      : JSON.parse(readFileSync(filePath, "utf-8"));

    if (pendingTransactionsToSync.length === 0) {
      if (sourceNode) {
        this.log(`Delta between source and target pending transactions is 0.`);
      } else {
        this.log(`No txs found from file ${filePath}`);
      }
      return;
    }

    this.log(`Pending transactions to process: ${pendingTransactionsToSync.length}`);

    // track errors serializing transactions
    let errorSerialization = 0;
    const transactions: Array<string> = [];

    for (const tx of pendingTransactionsToSync) {
      const transaction: UnsignedTransaction = {
        to: tx.to,
        nonce: Number.parseInt(tx.nonce.toString()),
        gasLimit: BigNumber.from(tx.gas),
        ...(Number(tx.type) === 2
          ? {
              gasPrice: BigNumber.from(tx.maxFeePerGas),
              maxFeePerGas: BigNumber.from(tx.maxFeePerGas),
              maxPriorityFeePerGas: BigNumber.from(tx.maxPriorityFeePerGas),
            }
          : { gasPrice: BigNumber.from(tx.gasPrice) }),
        data: tx.input || "0x",
        value: BigNumber.from(tx.value),
        ...(tx.chainId && Number(tx.type) !== 0 ? { chainId: Number.parseInt(tx.chainId.toString()) } : {}),
        ...(Number(tx.type) === 1 || Number(tx.type) === 2 ? { accessList: tx.accessList, type: Number(tx.type) } : {}),
      };

      const rawTx = ethers.utils.serializeTransaction(
        transaction,
        ethers.utils.splitSignature({
          // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
          v: Number.parseInt(tx.v!.toString()),
          // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
          r: tx.r!,
          s: tx.s,
        }),
      );

      if (ethers.utils.keccak256(rawTx) !== tx.hash) {
        errorSerialization++;
        this.warn(`Failed to serialize transaction: ${tx.hash}`);
      }
      transactions.push(rawTx);
    }

    const totalBatchesToProcess = Math.ceil(transactions.length / concurrentCount);

    this.log(`Total serialization errors: ${errorSerialization}`);

    let totalSuccess = 0;
    let success = 0;
    let errors = 0;
    let totalErrors = 0;

    if (flags["dry-run"]) {
      this.log(`Total batches to process: ${totalBatchesToProcess}`);
      return;
    }

    for (let i = 0; i < totalBatchesToProcess; i++) {
      const batchIndex = concurrentCount * i;
      const transactionBatch = transactions.slice(batchIndex, batchIndex + concurrentCount);

      if (transactionBatch.length === 0) {
        break;
      }

      this.log(`Processing batch: ${i + 1} of ${totalBatchesToProcess}, size ${transactionBatch.length}`);

      const transactionPromises = transactionBatch.map((transactionReq) => {
        return targetProvider.sendTransaction(transactionReq);
      });

      const results = await Promise.allSettled(transactionPromises);
      success += results.filter((result: PromiseSettledResult<unknown>) => result.status === "fulfilled").length;

      const resultErrors = results.filter(
        (result: PromiseSettledResult<unknown>): result is PromiseRejectedResult => result.status === "rejected",
      );
      errors += resultErrors.length;

      resultErrors.forEach((result) => {
        this.log(`${result.reason.message.toString()}`);
      });

      totalSuccess += success;
      totalErrors += errors;

      this.log(
        `
        Total count: ${transactionBatch.length + batchIndex} - Success: ${success} - Errors: ${errors} - Total Success: ${totalSuccess} - Total Errors: ${totalErrors}
      `.trim(),
      );

      success = 0;
      errors = 0;
    }
  }
}
