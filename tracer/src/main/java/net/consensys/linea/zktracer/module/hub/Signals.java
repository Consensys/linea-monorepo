/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.module.hub;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.opcode.InstructionFamily;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;

/**
 * Encodes the signals triggering other components.
 *
 * <p>When a component is requested, also checks that it may actually be triggered in the current
 * context.
 */
@Accessors(fluent = true)
public class Signals {
  private final Hub hub;
  private final OpCodeData opCodeData;

  @Getter boolean mmu = false;
  @Getter boolean mxp = false;
  @Getter boolean oob = false;
  @Getter boolean precompileInfo = false;
  @Getter boolean stipend = false;
  @Getter boolean exp = false;

  public Signals() {
    this.hub = null;
    this.opCodeData = null;
  }

  public Signals(Hub hub) {
    this.hub = hub;
    this.opCodeData = hub.currentFrame().opCode().getData();
  }

  public Signals snapshot() {
    var r = new Signals();
    r.mmu = this.mmu;
    r.mxp = this.mxp;
    r.oob = this.oob;
    r.precompileInfo = this.precompileInfo;
    r.stipend = this.stipend;
    r.exp = this.exp;
    return r;
  }

  /**
   * If requested and possible, trigger the MMU
   *
   * @return the signals
   */
  public Signals wantMmu(boolean b) {
    this.mmu = b && canMmu();
    return this;
  }

  /**
   * If requested and possible, trigger the MXP
   *
   * @return the signals
   */
  public Signals wantMxp(boolean b) {
    this.mxp = b && canMxp();
    return this;
  }

  /**
   * If requested and possible, trigger the OOB
   *
   * @return the signals
   */
  public Signals wantOob(boolean b) {
    this.oob = b && canOob();
    return this;
  }

  /**
   * If requested and possible, trigger the PREC_INFO
   *
   * @return the signals
   */
  public Signals wantPrecompileInfo(boolean b) {
    this.precompileInfo = b && canPrecompileInfo();
    return this;
  }

  /**
   * If requested and possible, trigger the STP
   *
   * @return the signals
   */
  public Signals wantStipend(boolean b) {
    this.stipend = b && canStp();
    return this;
  }

  /**
   * If requested and possible, trigger the EXP
   *
   * @return the signals
   */
  public Signals wantExp(boolean b) {
    this.exp = b && canExp();
    return this;
  }

  /**
   * If possible, trigger the MMU
   *
   * @return the signals
   */
  public Signals wantMmu() {
    this.mmu = canMmu();
    return this;
  }

  /**
   * If possible, trigger the MXP
   *
   * @return the signals
   */
  public Signals wantMxp() {
    this.mxp = canMxp();
    return this;
  }

  /**
   * If possible, trigger the OOB
   *
   * @return the signals
   */
  public Signals wantOob() {
    this.oob = canOob();
    return this;
  }

  /**
   * If possible, trigger the PREC_INFO
   *
   * @return the signals
   */
  public Signals wantPrecompileInfo() {
    this.precompileInfo = canPrecompileInfo();
    return this;
  }

  /**
   * If possible, trigger the STP
   *
   * @return the signals
   */
  public Signals wantStipend() {
    this.stipend = canStp();
    return this;
  }

  /**
   * If possible, trigger the EXP
   *
   * @return the signals
   */
  public Signals wantExp() {
    this.exp = canExp();
    return this;
  }

  private boolean canMmu() {
    // TODO: not enabled in the instruction decoding
    // return opCodeData.ramSettings().enabled() &&
    return hub.exceptions().none();
  }

  private boolean canMxp() {
    boolean mxpFlag = false;
    if (opCodeData.isMxp()) {
      mxpFlag =
          switch (opCodeData.instructionFamily()) {
            case CALL, CREATE, LOG -> !hub.exceptions().stackOverflow()
                || !hub.exceptions().staticViolation();
            default -> false;
          };
    }
    return mxpFlag;
  }

  private boolean canOob() {
    boolean oobFlag = false;

    if (opCodeData.stackSettings().oobFlag()) {
      if (opCodeData.mnemonic() == OpCode.CALLDATALOAD) {
        oobFlag = !hub.exceptions().stackUnderflow();
      } else if (opCodeData.instructionFamily() == InstructionFamily.JUMP) {
        oobFlag = !hub.exceptions().stackUnderflow();
      } else if (opCodeData.mnemonic() == OpCode.RETURNDATACOPY) {
        oobFlag = !hub.exceptions().stackUnderflow(); // TODO: updateCallerReturndata it
      } else if (opCodeData.instructionFamily() == InstructionFamily.CALL) {
        oobFlag = !hub.exceptions().any();
      } else if (opCodeData.instructionFamily() == InstructionFamily.CREATE) {
        oobFlag = !hub.exceptions().any();
      } else if (opCodeData.mnemonic() == OpCode.SSTORE) {
        oobFlag = !hub.exceptions().stackUnderflow() && !hub.exceptions().staticViolation();
      } else if (opCodeData.mnemonic() == OpCode.RETURN) {
        oobFlag =
            !hub.exceptions().stackUnderflow()
                && hub.currentFrame().underDeployment(); // TODO: see for the rest
      }
    }

    return oobFlag;
  }

  private boolean canPrecompileInfo() {
    boolean precompileInfoFlag = false;

    if (opCodeData.instructionFamily() == InstructionFamily.CALL) {
      precompileInfoFlag =
          !hub.exceptions()
              .any(); // TODO:  && no abort(assez de balance && CSD < 1024) && to precompile
    }

    return precompileInfoFlag;
  }

  private boolean canStp() {
    boolean stpFlag = false;

    if (opCodeData.instructionFamily() == InstructionFamily.CALL) {
      stpFlag =
          !hub.exceptions().stackUnderflow()
              && !hub.exceptions().staticViolation()
              && !hub.exceptions().outOfMemoryExpansion();
    } else if (opCodeData.instructionFamily() == InstructionFamily.CREATE) {
      stpFlag = false; // TODO:
    }

    return stpFlag;
  }

  private boolean canExp() {
    return opCodeData.mnemonic() == OpCode.EXP && !hub.exceptions().stackUnderflow();
  }
}
