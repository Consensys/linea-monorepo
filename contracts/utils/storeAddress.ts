/* eslint-disable @typescript-eslint/no-var-requires */
const editJsonFile = require("edit-json-file");

export const storeAddress = async (contractName: string, address: string, networkName: string) => {
  const file = editJsonFile(`./deployments.json`);
  file.set([networkName], networkName);
  file.set([networkName] + "." + [contractName], address);
  file.save();
};

export const delay = (ms: number) => new Promise((res) => setTimeout(res, ms));
