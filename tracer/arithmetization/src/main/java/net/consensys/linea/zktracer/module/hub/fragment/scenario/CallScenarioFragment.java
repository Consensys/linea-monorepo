/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package net.consensys.linea.zktracer.module.hub.fragment.scenario;

import java.util.List;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;

public class CallScenarioFragment implements TraceFragment {

  @Setter @Getter CallScenario scenario;

  public CallScenarioFragment() {
    this.scenario = CallScenario.UNDEFINED;
  }

  public CallScenarioFragment(final CallScenario scenario) {
    this.scenario = scenario;
  }

  public enum CallScenario {
    UNDEFINED,
    CALL_EXCEPTION,
    CALL_ABORT_WILL_REVERT,
    CALL_ABORT_WONT_REVERT,
    // Externally owned account call scenarios
    CALL_EOA_SUCCESS_WILL_REVERT,
    CALL_EOA_SUCCESS_WONT_REVERT,
    // Smart contract call scenarios:
    CALL_SMC_UNDEFINED,
    CALL_SMC_FAILURE_WILL_REVERT,
    CALL_SMC_FAILURE_WONT_REVERT,
    CALL_SMC_SUCCESS_WILL_REVERT,
    CALL_SMC_SUCCESS_WONT_REVERT,
    // Precompile call scenarios:
    CALL_PRC_UNDEFINED,
    CALL_PRC_FAILURE,
    CALL_PRC_SUCCESS_WILL_REVERT,
    CALL_PRC_SUCCESS_WONT_REVERT;

    public boolean isPrecompileScenario() {
      return this == CALL_PRC_FAILURE
          || this == CALL_PRC_SUCCESS_WILL_REVERT
          || this == CALL_PRC_SUCCESS_WONT_REVERT;
    }

    public boolean noLongerUndefined() {
      return this != UNDEFINED && this != CALL_PRC_UNDEFINED && this != CALL_SMC_UNDEFINED;
    }
  }

  private static final List<CallScenario> illegalTracingScenario =
      List.of(
          CallScenario.UNDEFINED, CallScenario.CALL_SMC_UNDEFINED, CallScenario.CALL_PRC_UNDEFINED);

  public Trace trace(Trace trace) {
    Preconditions.checkArgument(
        this.scenario.noLongerUndefined(), "Final Scenario hasn't been set");
    return trace
        .peekAtScenario(true)
        // // CALL scenarios
        ////////////////////
        .pScenarioCallException(this.scenario.equals(CallScenario.CALL_EXCEPTION))
        .pScenarioCallAbortWillRevert(this.scenario.equals(CallScenario.CALL_ABORT_WILL_REVERT))
        .pScenarioCallAbortWontRevert(this.scenario.equals(CallScenario.CALL_ABORT_WONT_REVERT))
        .pScenarioCallPrcFailure(this.scenario.equals(CallScenario.CALL_PRC_FAILURE))
        .pScenarioCallPrcSuccessCallerWillRevert(
            this.scenario.equals(CallScenario.CALL_PRC_SUCCESS_WILL_REVERT))
        .pScenarioCallPrcSuccessCallerWontRevert(
            this.scenario.equals(CallScenario.CALL_PRC_SUCCESS_WONT_REVERT))
        .pScenarioCallSmcFailureCallerWillRevert(
            this.scenario.equals(CallScenario.CALL_SMC_FAILURE_WILL_REVERT))
        .pScenarioCallSmcFailureCallerWontRevert(
            this.scenario.equals(CallScenario.CALL_SMC_FAILURE_WONT_REVERT))
        .pScenarioCallSmcSuccessCallerWillRevert(
            this.scenario.equals(CallScenario.CALL_SMC_SUCCESS_WILL_REVERT))
        .pScenarioCallSmcSuccessCallerWontRevert(
            this.scenario.equals(CallScenario.CALL_SMC_SUCCESS_WONT_REVERT))
        .pScenarioCallEoaSuccessCallerWillRevert(
            this.scenario.equals(CallScenario.CALL_EOA_SUCCESS_WILL_REVERT))
        .pScenarioCallEoaSuccessCallerWontRevert(
            this.scenario.equals(CallScenario.CALL_EOA_SUCCESS_WONT_REVERT));
  }
}
