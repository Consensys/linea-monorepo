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

package net.consensys.linea.zktracer.module.mmu;

import lombok.Builder;

@Builder
record ReadPad(int totalNumberLimbs, int totalNumberPaddingMicroInstructions) {

  int totalNumber() {
    return totalNumberLimbs + totalNumberPaddingMicroInstructions;
  }

  boolean isRead(final int processingRow) {
    return processingRow < totalNumberLimbs;
  }

  boolean isPad(final int processingRow) {
    return !isRead(processingRow);
  }

  boolean isFirstRead(final int processingRow) {
    return processingRow == 0 && totalNumberLimbs != 0;
  }

  boolean isFirstPad(final int processingRow) {
    return totalNumberPaddingMicroInstructions != 0 && processingRow == totalNumberLimbs;
  }

  boolean isFirstMicroInstruction(final int processingRow) {
    return isFirstRead(processingRow) || (totalNumberLimbs == 0 && isFirstPad(processingRow));
  }

  boolean isLastRead(final int processingRow) {
    return totalNumberLimbs != 0 && (processingRow + 1) == totalNumberLimbs;
  }

  boolean isLastPad(final int processingRow) {
    return totalNumberPaddingMicroInstructions != 0 && processingRow == (totalNumber() - 1);
  }

  int remainingMicroInstructions(final int processingRow) {
    if (!isRead(processingRow) || isPad(processingRow)) {
      return totalNumber();
    }

    return totalNumber() - processingRow - 1;
  }

  int remainingReads(final int processingRow) {
    if (isRead(processingRow)) {
      return totalNumberLimbs - processingRow - 1;
    }

    return 0;
  }

  int remainingPads(final int processingRow) {
    if (processingRow == -1 || isRead(processingRow)) {
      return totalNumberPaddingMicroInstructions;
    }

    return totalNumberPaddingMicroInstructions
        - (processingRow - totalNumberPaddingMicroInstructions)
        - 1;
  }
}
