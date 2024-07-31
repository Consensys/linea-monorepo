const eventMessageHashes = {
  anonymous: false,
  inputs: [
    {
      indexed: false,
      internalType: "bytes32[]",
      name: "messageHashes",
      type: "bytes32[]",
    },
  ],
  name: "L1L2MessageHashesAddedToInbox",
  type: "event",
} as const;

export default eventMessageHashes;
