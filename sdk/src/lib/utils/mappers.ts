import { LogDescription } from "ethers/lib/utils";
import { Message } from "./types";
import { MessageSentEvent } from "../../typechain/ZkEvmV2";

export const mapMessageSentEventOrLogToMessage = (event: MessageSentEvent | LogDescription): Message => {
  const { _from, _to, _fee, _value, _nonce, _calldata, _messageHash } = event.args;
  return {
    messageSender: _from,
    destination: _to,
    fee: _fee,
    value: _value,
    messageNonce: _nonce,
    calldata: _calldata,
    messageHash: _messageHash,
  };
};
