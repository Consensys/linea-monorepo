import { IContractClientLibrary } from "ts-libs/linea-shared-utils/src/core/client/IContractClientLibrary";
import { IYieldManager } from "../core/services/contracts/IYieldManager";
import { Address, BaseError, PublicClient, TransactionReceipt } from "viem";
import YieldManagerABI from "../core/abis/YieldManager.abi";

export class YieldManagerContractClient implements IYieldManager<TransactionReceipt> {
  private readonly blockchainClient: PublicClient;

  constructor(
    contractClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt, BaseError>,
    private readonly contractAddress: Address,
  ) {
    this.blockchainClient = contractClientLibrary.getBlockchainClient();
  }

  async getTargetReserveDeficit(_yieldProvider: string): Promise<bigint> {
    return this.blockchainClient.readContract({
      address: this.contractAddress,
      abi: YieldManagerABI,
      functionName: "getTargetReserveDeficit",
    });
  }

  async isStakingPaused(yieldProvider: string): Promise<boolean> {
    return this.blockchainClient.readContract({
      address: this.contractAddress,
      abi: YieldManagerABI,
      functionName: "isStakingPaused",
      args: [yieldProvider as Address],
    });
  }

  async isOssificationInitiated(yieldProvider: string): Promise<boolean> {
    return this.blockchainClient.readContract({
      address: this.contractAddress,
      abi: YieldManagerABI,
      functionName: "isOssificationInitiated",
      args: [yieldProvider as Address],
    });
  }

  async isOssified(yieldProvider: string): Promise<boolean> {
    return this.blockchainClient.readContract({
      address: this.contractAddress,
      abi: YieldManagerABI,
      functionName: "isOssified",
      args: [yieldProvider as Address],
    });
  }

  async withdrawableValue(yieldProvider: string): Promise<bigint> {
    return this.blockchainClient.readContract({
      address: this.contractAddress,
      abi: YieldManagerABI,
      functionName: "withdrawableValue",
      args: [yieldProvider as Address],
    });
  }
}

// export async function getLastInvoiceDate(client: Client, contractAddress: Address): Promise<Result<bigint, BaseError>> {
//   try {
//     const lastInvoiceDate = await readContract(client, {
//       address: contractAddress,
//       abi: [
//         {
//           inputs: [],
//           name: "lastInvoiceDate",
//           outputs: [
//             {
//               internalType: "uint256",
//               name: "",
//               type: "uint256",
//             },
//           ],
//           stateMutability: "view",
//           type: "function",
//         },
//       ],
//       functionName: "lastInvoiceDate",
//     });
//     return ok(lastInvoiceDate);
//   } catch (error) {
//     if (error instanceof BaseError) {
//       const decodedError = error.walk();
//       return err(decodedError as BaseError);
//     }
//     return err(error as BaseError);
//   }
// }

// const contract = getContract({ address, abi, client: publicClient })

// // The below will send a single request to the RPC Provider.
// const [name, totalSupply, symbol, balance] = await Promise.all([
//   contract.read.name(),
//   contract.read.totalSupply(),
//   contract.read.symbol(),
//   contract.read.balanceOf([address]),
// ])
