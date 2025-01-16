import { ethers } from "ethers";
import yargs from "yargs";
import { hideBin } from "yargs/helpers";
import { GoNativeCompressor } from "../compressor/GoNativeCompressor";

const argv = yargs(hideBin(process.argv))
  .option("raw-tx", {
    describe: "Rlp encoded transaction",
    type: "string",
    demandOption: true,
  })
  .parseSync();

async function main(args: typeof argv) {
  const dataLimit = 800_000;
  const compressor = new GoNativeCompressor(dataLimit);

  const { rawTx } = args;

  const size = calculateTransactionSizeFromRlPEncodedTx(compressor, rawTx);

  console.log(`Transaction size: ${size}`);
}

function calculateTransactionSizeFromRlPEncodedTx(compressor: GoNativeCompressor, tx: string): number {
  try {
    const rlpEncodedTx = ethers.encodeRlp(tx);
    console.log(rlpEncodedTx)
    const rlpEncodedTxInBytes = ethers.getBytes(rlpEncodedTx);
    return compressor.getCompressedTxSize(rlpEncodedTxInBytes);
  } catch (error) {
    throw new Error(`Transaction size calculation error: ${error}`);
  }
}

main(argv)
  .then(() => {
    process.exit(0);
  })
  .catch((e) => {
    console.error(e);
    process.exit(1);
  });







