/* Sends ETH from a Web3Signer-managed account to a destination address.

Uses the Web3SignerClientAdapter for mTLS authentication and transaction signing.
See README.md in this directory for full setup instructions including trust store generation.

Usage (from repo root):

1. Start the full local stack (L1 + L2):

make start-env

2. Start Web3Signer with TLS (in a separate terminal):

docker run --rm \
  --platform=linux/amd64 \
  --name web3signer \
  -p 9000:9000 \
  -v "$(pwd)/docker/web3signer/key-files:/key-files" \
  -v "$(pwd)/docker/web3signer/tls-files:/tls-files" \
  consensys/web3signer:25.12.0 \
  --key-store-path=/key-files/ \
  --tls-keystore-file=/tls-files/web3signer-keystore.p12 \
  --tls-keystore-password-file=/tls-files/web3signer-keystore-password.txt \
  --tls-known-clients-file=/tls-files/known-clients.txt \
  --http-host-allowlist='*' \
  eth1 \
  --chain-id=31648428

3. Run the script:

RPC_URL=http://localhost:8445 \
WEB3_SIGNER_URL=https://127.0.0.1:9000 \
WEB3_SIGNER_PUBLIC_KEY=ba5734d8f7091719471e7f7ed6b9df170dc70cc661ca05e688601ad984f068b0d67351e5f06073092499336ab0839ef8a521afd334e53807205fa2f08eec74f4 \
WEB3_SIGNER_KEYSTORE_PATH="$(pwd)/docker/config/linea-besu-sequencer/tls-files/sequencer_client_keystore.p12" \
WEB3_SIGNER_KEYSTORE_PASSPHRASE=changeit \
WEB3_SIGNER_TRUST_STORE_PATH="$(pwd)/docker/config/linea-besu-sequencer/tls-files/web3signer_truststore.p12" \
WEB3_SIGNER_TRUST_STORE_PASSPHRASE=changeit \
DESTINATION_ADDRESS=0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 \
SEND_VALUE_WEI=1000000000000000000 \
pnpm --filter @consensys/linea-shared-utils exec tsx scripts/send-eth-via-web3-signer.ts

Available Web3Signer accounts (local dev — DO NOT REUSE ELSEWHERE):

  finalization-signer
    Public key: ba5734d8f7091719471e7f7ed6b9df170dc70cc661ca05e688601ad984f068b0d67351e5f06073092499336ab0839ef8a521afd334e53807205fa2f08eec74f4
    L1 address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8

  data-submission-signer
    Public key: 9d9031e97dd78ff8c15aa86939de9b1e791066a0224e331bc962a2099a7b1f0464b8bbafe1535f2301c72c2cb3535b172da30b02686ab0393d348614f157fbdb
    L1 address: 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC

  anchoring-signer
    Public key: 4a788ad6fa008beed58de6418369717d7492f37d173d70e2c26d9737e2c6eeae929452ef8602a19410844db3e200a0e73f5208fd76259a8766b73953fc3e7023

  liveness-signer
    Public key: 3aee43fafba7d1b83b23448df3c9876751feaa19b7b38be8ee6e277f2761b009737841d9a5278c8ec466655264cff58ad7921dd1687b8f5ff303e9b2fecebf30
 */

import {
  type Address,
  createPublicClient,
  formatEther,
  type Hex,
  http,
  isAddress,
  parseSignature,
  serializeTransaction,
  type TransactionSerializableEIP1559,
} from "viem";
import { sendRawTransaction, waitForTransactionReceipt } from "viem/actions";

import { Web3SignerClientAdapter } from "../src/clients/Web3SignerClientAdapter";
import { WinstonLogger } from "../src/logging/WinstonLogger";

const REQUIRED_ENV_VARS = [
  "RPC_URL",
  "WEB3_SIGNER_URL",
  "WEB3_SIGNER_PUBLIC_KEY",
  "WEB3_SIGNER_KEYSTORE_PATH",
  "WEB3_SIGNER_KEYSTORE_PASSPHRASE",
  "WEB3_SIGNER_TRUST_STORE_PATH",
  "WEB3_SIGNER_TRUST_STORE_PASSPHRASE",
  "DESTINATION_ADDRESS",
  "SEND_VALUE_WEI",
] as const;

async function main() {
  const missing = REQUIRED_ENV_VARS.filter((name) => !process.env[name]);
  if (missing.length > 0) {
    throw new Error(`Missing required env vars: ${missing.join(", ")}`);
  }

  const rpcUrl = process.env.RPC_URL!;
  const destinationAddress = process.env.DESTINATION_ADDRESS! as Address;
  const valueWei = BigInt(process.env.SEND_VALUE_WEI!);

  if (!isAddress(destinationAddress)) {
    throw new Error(`Invalid destination address: ${destinationAddress}`);
  }

  const logger = new WinstonLogger("send-eth-via-web3-signer");
  const signer = new Web3SignerClientAdapter(
    logger,
    process.env.WEB3_SIGNER_URL!,
    process.env.WEB3_SIGNER_PUBLIC_KEY! as Hex,
    process.env.WEB3_SIGNER_KEYSTORE_PATH!,
    process.env.WEB3_SIGNER_KEYSTORE_PASSPHRASE!,
    process.env.WEB3_SIGNER_TRUST_STORE_PATH!,
    process.env.WEB3_SIGNER_TRUST_STORE_PASSPHRASE!,
  );

  const senderAddress = signer.getAddress();
  console.log(`Sender:      ${senderAddress}`);
  console.log(`Destination: ${destinationAddress}`);
  console.log(`Value:       ${valueWei} wei (${formatEther(valueWei)} ETH)`);

  const publicClient = createPublicClient({ transport: http(rpcUrl) });

  const [chainId, nonce, fees, balance] = await Promise.all([
    publicClient.getChainId(),
    publicClient.getTransactionCount({ address: senderAddress }),
    publicClient.estimateFeesPerGas(),
    publicClient.getBalance({ address: senderAddress }),
  ]);

  console.log(`Chain ID:    ${chainId}`);
  console.log(`Balance:     ${formatEther(balance)} ETH`);
  console.log(`Nonce:       ${nonce}`);

  if (fees.maxFeePerGas == null || fees.maxPriorityFeePerGas == null) {
    throw new Error("EIP-1559 fee data unavailable — chain may not support EIP-1559");
  }

  if (balance < valueWei) {
    throw new Error(
      `Insufficient balance: ${formatEther(balance)} ETH available, ` + `${formatEther(valueWei)} ETH requested`,
    );
  }

  const tx: TransactionSerializableEIP1559 = {
    type: "eip1559",
    chainId,
    nonce,
    to: destinationAddress,
    value: valueWei,
    gas: 21_000n,
    maxFeePerGas: fees.maxFeePerGas,
    maxPriorityFeePerGas: fees.maxPriorityFeePerGas,
  };

  console.log("\nSigning transaction...");
  const signature = await signer.sign(tx);
  const serializedTx = serializeTransaction(tx, parseSignature(signature));

  console.log("Broadcasting...");
  const txHash = await sendRawTransaction(publicClient, {
    serializedTransaction: serializedTx,
  });
  console.log(`TX hash:     ${txHash}`);

  console.log("Waiting for confirmation...");
  const receipt = await waitForTransactionReceipt(publicClient, {
    hash: txHash,
  });
  console.log(`Confirmed in block ${receipt.blockNumber} (gas used: ${receipt.gasUsed})`);
}

main().catch((err) => {
  console.error("Error:", err instanceof Error ? err.message : err);
  process.exit(1);
});
