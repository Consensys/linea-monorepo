import type { NewTaskActionFunction } from "hardhat/types/tasks";
import { getTaskCliOrEnvValue } from "../../../../common/helpers/environmentHelper.js";

interface TaskArgs {
  lineaRollup?: string;
  yieldManager?: string;
  yieldProvider?: string;
}

const action: NewTaskActionFunction<TaskArgs> = async (taskArgs) => {
  const lineaRollup = getTaskCliOrEnvValue(taskArgs, "lineaRollup", "LINEA_ROLLUP_ADDRESS");
  const yieldManager = getTaskCliOrEnvValue(taskArgs, "yieldManager", "YIELD_MANAGER_ADDRESS");
  const yieldProvider = getTaskCliOrEnvValue(taskArgs, "yieldProvider", "YIELD_PROVIDER_ADDRESS");

  if (!lineaRollup || !yieldManager || !yieldProvider) {
    throw new Error("Missing required params: lineaRollup, yieldManager, yieldProvider");
  }

  console.log("This is a testing task for unstakePermissionless. Implement as needed.");
  console.log("LineaRollup:", lineaRollup);
  console.log("YieldManager:", yieldManager);
  console.log("YieldProvider:", yieldProvider);
};

export default action;
