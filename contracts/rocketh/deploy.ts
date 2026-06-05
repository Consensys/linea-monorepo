import { setupDeployScripts } from "rocketh";

import { clearHandoffStore } from "../common/helpers/deploymentHandoff";

const { deployScript: rockethDeployScript } = setupDeployScripts({});
let handoffCleared = false;

function deployScript(
  callback: Parameters<typeof rockethDeployScript>[0],
  options: Parameters<typeof rockethDeployScript>[1],
): ReturnType<typeof rockethDeployScript> {
  return rockethDeployScript(async (env, args) => {
    if (!handoffCleared) {
      clearHandoffStore();
      handoffCleared = true;
    }

    return await callback(env, args);
  }, options);
}

export { deployScript };
