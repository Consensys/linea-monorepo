export const name = 'Linea';
export const version = '0.0.1';
export const license = 'MIT';

const WETH_ADDR = '0xe5d7c2a44ffddf6b295a15c148167daaaf5cf34f';
const USDC_ADDR = '0x176211869cA2b568f2A7D4EE941E073a821EE1ff';
const FROM_ADDR = '0x8C8766c1Ac7308604C80387f1bF8128386b64de9';
const TO_ADDR = '0xB1c25ff9F709cA3cd88D398586dd02aC62fce5BB'

const PANCAKE_SWAP_ROUTER = '0x1b81D678ffb9C0263b24A97847620C99d213eB14';
type TxType = 'EthTransfer' | 'ERC20Transfer' | 'ERC20Swap';
const PANCAKE_SWAP_EXACT_INPUT_SINGLE_ABI = [
  {
    "inputs": [
      {
        "components": [
          {
            "internalType": "address",
            "name": "tokenIn",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "tokenOut",
            "type": "address"
          },
          {
            "internalType": "uint24",
            "name": "fee",
            "type": "uint24"
          },
          {
            "internalType": "address",
            "name": "recipient",
            "type": "address"
          },
          {
            "internalType": "uint256",
            "name": "deadline",
            "type": "uint256"
          },
          {
            "internalType": "uint256",
            "name": "amountIn",
            "type": "uint256"
          },
          {
            "internalType": "uint256",
            "name": "amountOutMinimum",
            "type": "uint256"
          },
          {
            "internalType": "uint160",
            "name": "sqrtPriceLimitX96",
            "type": "uint160"
          }
        ],
        "internalType": "struct ISwapRouter.ExactInputSingleParams",
        "name": "params",
        "type": "tuple"
      }
    ],
    "name": "exactInputSingle",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "amountOut",
        "type": "uint256"
      }
    ],
    "stateMutability": "payable",
    "type": "function"
  }
]

const USDC_CONTRACT_TRANSFER_ABI = [
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "to",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "value",
        "type": "uint256"
      }
    ],
    "name": "transfer",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  }
]

export function setup(sdk: Context) {
  sdk.ethers.addProvider('linea', 'https://rpc.linea.build');

  const getFeeForCost = async (gasAmt: number) => {
    const weiPerGas = await sdk.ethers.getProvider('linea').getGasPrice();
    const ethPrice = await sdk.defiLlama.getCurrentPrice('coingecko', 'ethereum');

    return (weiPerGas * gasAmt * ethPrice) / 1e18;
  };

  const getGasAmount = async (txType: TxType): Promise<number> => {
    const provider = sdk.ethers.getProvider('linea');
    const pancakeSwapContract = sdk.ethers.getContract(PANCAKE_SWAP_ROUTER, PANCAKE_SWAP_EXACT_INPUT_SINGLE_ABI, 'linea');
    const usdcContract = sdk.ethers.getContract(USDC_ADDR, USDC_CONTRACT_TRANSFER_ABI, 'linea');

    if (txType === 'EthTransfer') {
      return provider.estimateGas({
        from: FROM_ADDR,
        to: TO_ADDR,
        value: '0x1',
        data: '0x'
      });
    } else if (txType === 'ERC20Transfer') {
      return usdcContract.estimateGas.transfer(TO_ADDR, 1, { from: FROM_ADDR });
    } else if (txType === 'ERC20Swap') {
      const params = {
        tokenIn: USDC_ADDR,
        tokenOut: WETH_ADDR,
        fee: 500,
        recipient: FROM_ADDR,
        deadline: 4102401129,
        amountIn: 9447000,
        amountOutMinimum: 0,
        sqrtPriceLimitX96: 0,
      }
      // swap 1 USDC to ETH using KyberSwap with transaction deadline of Dec 31, 2099
      return pancakeSwapContract.estimateGas.exactInputSingle(params, {
        from: FROM_ADDR,
      });
    }
  };

  sdk.register({
    id: 'linea',
    queries: {
      feeTransferEth: async () => getFeeForCost(await getGasAmount('EthTransfer')),
      feeTransferERC20: async () => getFeeForCost(await getGasAmount('ERC20Transfer')),
      feeSwap: async () => getFeeForCost(await getGasAmount('ERC20Swap')),
    },
    metadata: {
      icon: sdk.ipfs.getDataURILoader('QmXNhh135wo6jUuHdkhUTTV4p63sMpichCERuiyUDPe5vG', 'image/svg+xml'),
      category: 'l2',
      name: 'Linea',
      description: 'Linea zkEVM is the leading ZK scaling solution that is equivalent to Ethereum Virtual Machine.',
      l2BeatSlug: 'linea',
      website: 'https://linea.build',
    },
  });
}
