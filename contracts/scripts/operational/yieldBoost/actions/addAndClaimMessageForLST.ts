import type { NewTaskActionFunction } from "hardhat/types/tasks";
import { getTaskCliOrEnvValue } from "../../../../common/helpers/environmentHelper.js";

interface TaskArgs {
  lineaRollup?: string;
  yieldManager?: string;
  amount?: string;
}

const action: NewTaskActionFunction<TaskArgs> = async (taskArgs) => {
  const lineaRollup = getTaskCliOrEnvValue(taskArgs, "lineaRollup", "LINEA_ROLLUP_ADDRESS");
  const yieldManager = getTaskCliOrEnvValue(taskArgs, "yieldManager", "YIELD_MANAGER_ADDRESS");
  const amount = getTaskCliOrEnvValue(taskArgs, "amount", "AMOUNT");

  if (!lineaRollup || !yieldManager || !amount) {
    throw new Error("Missing required params: lineaRollup, yieldManager, amount");
  }

  console.log("This is a testing task for LST. Implement as needed.");
  console.log("LineaRollup:", lineaRollup);
  console.log("YieldManager:", yieldManager);
  console.log("Amount:", amount);
};

export default action;
