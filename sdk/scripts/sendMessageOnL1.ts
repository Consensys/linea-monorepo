import { BigNumber, BytesLike, ContractReceipt, PayableOverrides, Wallet } from "ethers";
import { config } from "dotenv";
import yargs from "yargs";
import { hideBin } from "yargs/helpers";
import { JsonRpcProvider } from "@ethersproject/providers";
import { L2MessageService__factory, ZkEvmV2__factory } from "../src/typechain/factories";
import { ZkEvmV2 } from "../src/typechain/ZkEvmV2";
import { SendMessageArgs } from "./types";
import { sanitizeAddress, sanitizePrivKey } from "./cli";
import { L2MessageService } from "../src/typechain";
import { encodeSendMessage } from "./helpers";

config();

const argv = yargs(hideBin(process.argv))
  .option("l1-rpc-url", {
    describe: "Rpc url",
    type: "string",
    demandOption: true,
  })
  .option("l2-rpc-url", {
    describe: "Rpc url",
    type: "string",
    demandOption: true,
  })
  .option("priv-key", {
    describe: "Signer private key",
    type: "string",
    demandOption: true,
    coerce: sanitizePrivKey("priv-key"),
  })
  .option("l1-contract-address", {
    describe: "Smart contract address",
    type: "string",
    demandOption: true,
    coerce: sanitizeAddress("smc-address"),
  })
  .option("l2-contract-address", {
    describe: "Smart contract address",
    type: "string",
    demandOption: true,
    coerce: sanitizeAddress("smc-address"),
  })
  .option("to", {
    describe: "Destination address",
    type: "string",
    demandOption: true,
    coerce: sanitizeAddress("to"),
  })
  .option("fee", {
    describe: "Fee passed to send message function",
    type: "number",
    demandOption: true,
  })
  .option("value", {
    describe: "Value passed to send message function",
    type: "number",
    demandOption: true,
  })
  .option("calldata", {
    describe: "Encoded message calldata",
    type: "string",
    demandOption: true,
  })
  .option("number-of-message", {
    describe: "Number of messages to send",
    type: "number",
    demandOption: true,
  })
  .parseSync();

const sendMessage = async (
  contract: ZkEvmV2,
  args: SendMessageArgs,
  overrides?: PayableOverrides,
): Promise<ContractReceipt> => {
  const tx = await contract.sendMessage(args.to, args.fee, args.calldata, overrides);
  return await tx.wait();
};

const sendMessages = async (
  contract: ZkEvmV2,
  numberOfMessages: number,
  args: SendMessageArgs,
  overrides?: PayableOverrides,
) => {
  let nonce = await contract.signer.getTransactionCount();
  const sendMessagePromises: Promise<ContractReceipt>[] = [];

  for (let i = 0; i < numberOfMessages; i++) {
    sendMessagePromises.push(
      sendMessage(contract, args, {
        ...overrides,
        nonce,
      }),
    );
    nonce++;
  }

  await Promise.all(sendMessagePromises);
};

const getMessageCounter = async (contractAddress: string, signer: Wallet) => {
  const zkEvmV2 = ZkEvmV2__factory.getContract(contractAddress, ZkEvmV2__factory.createInterface(), signer) as ZkEvmV2;
  return await zkEvmV2.nextMessageNumber();
};

const anchorMessageHashesOnL2 = async (contractAddress: string, signer: Wallet, messageHashes: BytesLike[]) => {
  const l2MessageService = L2MessageService__factory.getContract(
    contractAddress,
    L2MessageService__factory.createInterface(),
    signer,
  ) as L2MessageService;
  const tx = await l2MessageService.addL1L2MessageHashes(messageHashes);
  await tx.wait();
};

const main = async (args: typeof argv) => {
  const l1Provider = new JsonRpcProvider(args.l1RpcUrl);
  const l2Provider = new JsonRpcProvider(args.l2RpcUrl);
  const l1Signer = new Wallet(args.privKey, l1Provider);
  const l2Signer = new Wallet(args.privKey, l2Provider);

  const functionArgs: SendMessageArgs & PayableOverrides = {
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    to: args.to,
    fee: BigNumber.from(args.fee.toString()),
    calldata: args.calldata,
  };

  const zkEvmV2 = ZkEvmV2__factory.connect(args.l1ContractAddress, l1Signer) as ZkEvmV2;

  await sendMessages(zkEvmV2, args.numberOfMessage, functionArgs, { value: BigNumber.from(args.value.toString()) });

  // Anchor messages hash on L2
  const nextMessageCounter = await getMessageCounter(args.l1ContractAddress, l1Signer);
  const startCounter = nextMessageCounter.sub(args.numberOfMessage);

  const messageHashesToAnchor: string[] = [];
  for (let i = startCounter.toNumber(); i < nextMessageCounter.toNumber(); i++) {
    const messageHash = await encodeSendMessage(
      l1Signer.address,
      args.to,
      BigNumber.from(args.fee.toString()),
      BigNumber.from(args.value.toString()).sub(args.fee.toString()),
      BigNumber.from(i),
      args.calldata,
    );
    console.log(messageHash);
    messageHashesToAnchor.push(messageHash);
  }

  await anchorMessageHashesOnL2(args.l2ContractAddress, l2Signer, messageHashesToAnchor);
};

main(argv)
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
