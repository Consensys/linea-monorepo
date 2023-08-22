import { MessageEntity } from "../entity/Message.entity";
import { MessageInDb } from "./types";

export const mapMessageToMessageEntity = (message: MessageInDb): MessageEntity => {
  return {
    id: message?.id as number,
    ...message,
    createdAt: new Date(),
    updatedAt: new Date(),
  };
};

export const mapMessageEntityToMessage = (entity: MessageEntity): MessageInDb => {
  return {
    messageSender: entity.messageSender,
    destination: entity.destination,
    fee: entity.fee,
    value: entity.value,
    messageNonce: entity.messageNonce,
    calldata: entity.calldata,
    messageHash: entity.messageHash,
    messageContractAddress: entity.messageContractAddress,
    sentBlockNumber: entity.sentBlockNumber,
    direction: entity.direction,
    status: entity.status,
    claimTxCreationDate: entity.claimTxCreationDate,
    claimTxGasLimit: entity.claimTxGasLimit,
    claimTxMaxFeePerGas: entity.claimTxMaxFeePerGas,
    claimTxMaxPriorityFeePerGas: entity.claimTxMaxPriorityFeePerGas,
    claimTxNonce: entity.claimTxNonce,
    claimTxHash: entity.claimTxHash,
    claimNumberOfRetry: entity.claimNumberOfRetry,
    claimLastRetriedAt: entity.claimLastRetriedAt,
    claimGasEstimationThreshold: entity.claimGasEstimationThreshold,
    createdAt: entity.createdAt,
    updatedAt: entity.updatedAt,
  };
};
