import { IContractClientLibrary } from "../../core/client/IContractClientLibrary";
import { createPublicClient, http, PublicClient } from 'viem'
import { mainnet } from 'viem/chains'

// Re-use via composition in ContractClients
// Hope that using strategy pattern like this makes us more 'viem-agnostic'
export class EthereumMainnetClientLibrary implements IContractClientLibrary {
    blockchainReadClient: PublicClient;

    constructor(rpcUrl: string) {
        // Aim re-use single blockchain client for
        // i.) Better connection pooling
        // ii.) Memory efficient
        // iii.) Single point of configuration
        this.blockchainReadClient = createPublicClient({
            chain: mainnet,
            transport: http(rpcUrl, { batch: true, retryCount: 3 }),
        });
    }

    getBlockchainReadClient(): PublicClient {
        return this.blockchainReadClient;
    }

    estimateTransactionGas

    sendSerializedTransaction(tx: Hex)

    // const gasEstimationResult = await estimateTransactionGas(client, {
    //   to: contractAddress,
    //   account: senderAddress,
    //   value: 0n,
    //   data: encodeFunctionData({
    //     abi: SUBMIT_INVOICE_ABI,
    //     functionName: "submitInvoice",
    //     args: [
    //       BigInt(Math.floor(invoicePeriod.startDate.getTime() / 1000)),
    //       BigInt(Math.floor(invoicePeriod.endDate.getTime() / 1000)),
    //       BigInt(totalCostsInEth),
    //     ],
    //   }),
    // });
}
