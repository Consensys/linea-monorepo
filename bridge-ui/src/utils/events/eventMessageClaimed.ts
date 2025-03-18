const eventMessageClaimed = {
  anonymous: false,
  inputs: [{ indexed: true, internalType: "bytes32", name: "_messageHash", type: "bytes32" }],
  name: "MessageClaimed",
  type: "event",
} as const;

export default eventMessageClaimed;
