import { MessageSent } from "@consensys/linea-sdk";

export interface IMessageRetriever<TransactionReceipt> {
  getMessageByMessageHash(messageHash: string): Promise<MessageSent | null>;
  getMessagesByTransactionHash(transactionHash: string): Promise<MessageSent[] | null>;
  getTransactionReceiptByMessageHash(messageHash: string): Promise<TransactionReceipt | null>;
}
