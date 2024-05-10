import { defineChain } from 'viem'
 
export const lineaSepolia = defineChain({
    id: 59_141,
    name: 'Linea Sepolia Testnet',
    network: 'linea-sepolia',
    nativeCurrency: { name: 'Linea Ether', symbol: 'ETH', decimals: 18 },
    rpcUrls: {
      infura: {
        http: ['https://linea-sepolia.infura.io/v3'],
        webSocket: ['wss://linea-sepolia.infura.io/ws/v3'],
      },
      default: {
        http: ['https://rpc.sepolia.linea.build'],
        webSocket: ['wss://rpc.sepolia.linea.build'],
      },
      public: {
        http: ['https://rpc.sepolia.linea.build'],
        webSocket: ['wss://rpc.sepolia.linea.build'],
      },
    },
    blockExplorers: {
      default: {
        name: 'Etherscan',
        url: 'https://sepolia.lineascan.build',
      },
      etherscan: {
        name: 'Etherscan',
        url: 'https://sepolia.lineascan.build',
      },
      blockscout: {
        name: 'Blockscout',
        url: 'https://explorer.sepolia.linea.build',
      },
    },
    testnet: true,
})