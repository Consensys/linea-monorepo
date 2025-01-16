import { ethers } from "ethers";
import yargs from "yargs";
import { hideBin } from "yargs/helpers";
import { GoNativeCompressor } from "../compressor/GoNativeCompressor";

const argv = yargs(hideBin(process.argv))
  .option("rlp-encoded-tx", {
    describe: "Rlp encoded transaction",
    type: "string",
    demandOption: true,
  })
  .parseSync();

async function main(args: typeof argv) {
  const dataLimit = 800_000;
  const compressor = new GoNativeCompressor(dataLimit);

  const { rlpEncodedTx } = args;

  const size = calculateTransactionSizeFromRlPEncodedTx(compressor, rlpEncodedTx);

  console.log(`Transaction size: ${size}`);
}

function calculateTransactionSizeFromRlPEncodedTx(compressor: GoNativeCompressor, rlpEncodedTx: string): number {
  try {
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
