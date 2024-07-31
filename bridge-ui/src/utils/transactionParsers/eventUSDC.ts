const eventUSDC = {
  anonymous: false,
  inputs: [
    {
      indexed: true,
      internalType: "address",
      name: "depositor",
      type: "address",
    },
    {
      indexed: false,
      internalType: "uint256",
      name: "amount",
      type: "uint256",
    },
    { indexed: true, internalType: "address", name: "to", type: "address" },
  ],
  name: "Deposited",
  type: "event",
} as const;

export default eventUSDC;
