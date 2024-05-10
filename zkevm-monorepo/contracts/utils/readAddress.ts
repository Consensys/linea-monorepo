import fs from "fs";
import path from "path";

export const getDeployedContractOnNetwork = async (
  networkName: string,
  contractName: string,
): Promise<string | undefined> => {
  const filePath = path.join(__dirname, "..", "deployments", `${networkName}`, `${contractName}.json`);
  if (!fs.existsSync(filePath)) {
    return undefined;
  }
  const data = fs.readFileSync(filePath, "utf-8");
  return JSON.parse(data).address;
};

export const delay = (ms: number) => new Promise((res) => setTimeout(res, ms));
