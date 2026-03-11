import { IProvider } from "../../../core/clients/blockchain/IProvider";
import { BaseError } from "../../../core/errors";
import { IReceiptPoller } from "../../../core/services/IReceiptPoller";
import { Hash, TransactionReceipt } from "../../../core/types";
import { wait } from "../../../core/utils/shared";

export class ViemReceiptPoller implements IReceiptPoller {
  constructor(private readonly provider: IProvider) {}

  public async poll(transactionHash: Hash, timeout: number, interval: number): Promise<TransactionReceipt> {
    const deadline = Date.now() + timeout;
    while (Date.now() < deadline) {
      const receipt = await this.provider.getTransactionReceipt(transactionHash);
      if (receipt) return receipt;
      await wait(interval);
    }
    throw new BaseError(`Transaction receipt not found after polling. transactionHash=${transactionHash}`);
  }
}
