import { readFileSync, access, constants, watch } from "fs";
import { ethers } from "ethers";
import { getBlockchainNode } from "../common";
import { dirname, basename as _basename } from "path";

class LoggingProvider extends ethers.providers.JsonRpcProvider {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  async perform(method: string, parameters: any): Promise<any> {
    console.log(">>>", method, parameters);
    return super.perform(method, parameters).then((result) => {
      console.log("<<<", method, parameters, result);
      return result;
    });
  }
}

function getProvider() {
  const blockchainNode = getBlockchainNode();
  console.log(`Blockchain RPC at ${blockchainNode}`);

  // "If verbose output is needed please set VERBOSE_BLOCKCHAIN_LOG env"
  if (process.env.VERBOSE_BLOCKCHAIN_LOG) {
    return new LoggingProvider(blockchainNode);
  }
  return new ethers.providers.JsonRpcProvider(blockchainNode);
}

function getWallet(credentialsFile: string) {
  const credentials = JSON.parse(readFileSync(credentialsFile, "utf8"));
  const privateKey = "0x" + credentials.account_key.priv_key;
  const provider = getProvider();

  return new ethers.Wallet(privateKey, provider);
}

function getRollupContractData(rollupConfigPath: string) {
  const contractData = JSON.parse(readFileSync(rollupConfigPath, "utf8"));
  console.log(`Contract address: ${contractData.address}`);
  return contractData;
}

function getRollupContract(rollupConfigPath: string, wallet: ethers.Wallet) {
  const contractData = getRollupContractData(rollupConfigPath);
  return new ethers.Contract(contractData.address, contractData.abi, wallet);
}

// https://stackoverflow.com/a/47764403/995270
function waitUntilFileExists(filePath: string, timeout: number) {
  return new Promise<void>(function (resolve, reject) {
    const timer = setTimeout(function () {
      watcher.close();
      reject(new Error("File did not exists and was not created during the timeout."));
    }, timeout);

    access(filePath, constants.R_OK, function (err) {
      if (!err) {
        clearTimeout(timer);
        watcher.close();
        resolve();
      }
    });

    const dir = dirname(filePath);
    const basename = _basename(filePath);
    const watcher = watch(dir, function (eventType, filename) {
      if (eventType === "rename" && filename === basename) {
        clearTimeout(timer);
        watcher.close();
        resolve();
      }
    });
  });
}

/**
 * @param provider ethers Provider instance
 * @param percentile [0, 100] maxPriorityFeePerGas will be taken from the
 * previous block's reward distribution's given percentile bucket.
 * Used to define priority. 0 - minimal reward, inclusion not guaranteed, 100 -
 * maximum reward in the previous block, inclusion is very likely
 * @returns {Promise<{maxPriorityFeePerGas: *, maxFeePerGas: *}>}
 */
async function get1559Fees(
  provider: ethers.providers.JsonRpcProvider,
  percentile = 15,
): Promise<{ maxPriorityFeePerGas?: ethers.BigNumber; maxFeePerGas?: ethers.BigNumber }> {
  return provider.send("eth_feeHistory", ["0x1", "latest", [percentile]]).then((feeHistory) => {
    const maxPriorityFeePerGas = ethers.BigNumber.from(feeHistory.reward[0][0]);
    const maxFeePerGas = ethers.BigNumber.from(feeHistory.baseFeePerGas[feeHistory.baseFeePerGas.length - 1]).add(
      maxPriorityFeePerGas,
    );
    if (maxFeePerGas.gt(0) && maxPriorityFeePerGas.gt(0)) {
      return {
        maxFeePerGas: maxFeePerGas,
        maxPriorityFeePerGas: maxPriorityFeePerGas,
      };
    }
    return {};
  });
}

export {
  getProvider,
  getWallet,
  getRollupContract,
  getRollupContractData,
  waitUntilFileExists as checkExistsWithTimeout,
  get1559Fees,
};
