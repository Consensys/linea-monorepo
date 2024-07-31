/* eslint-disable @typescript-eslint/no-explicit-any */
import fs from "fs";
import path from "path";
import yargs from "yargs/yargs";
import { hideBin } from "yargs/helpers";

/*******************************USAGE******************************************************************
 ts-node ./scripts/gnosis/encodeProoflessOutput.ts --proofless-file "./prooflessExample.json"
*******************************************************************************************************/

const argv = yargs(hideBin(process.argv))
  .option("proofless-file", {
    describe: "Your proofless file name and location",
    type: "string",
    demandOption: true,
  })
  .parseSync();

// reading JSON file
function readJsonFile(filePath: string): unknown {
  const rawdata = fs.readFileSync(filePath);
  const jsonData = JSON.parse(rawdata.toString());
  return jsonData;
}

// formating data
function formatData(data: any): Array<any> {
  return [
    data.parentStateRootHash,
    data.dataHashes,
    data.dataParentHash,
    data.finalBlockNumber,
    data.lastFinalizedTimestamp,
    data.finalTimestamp,
    data.l1RollingHash,
    data.l1RollingHashMessageNumber,
    data.l2MerkleRoots,
    data.l2MerkleTreesDepth,
    data.l2MessagingBlocksOffsets,
  ];
}

async function main(args: typeof argv) {
  const filePath = path.join(__dirname, args.prooflessFile);
  const readJsonData = readJsonFile(filePath);
  const formattedFinalizationData = formatData(readJsonData);

  console.log("_finalizationData:");
  console.log(JSON.stringify(formattedFinalizationData));
  console.log("\n");
}

main(argv)
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
