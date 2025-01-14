/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.module.gas;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

/**
 * WARNING: This code is generated automatically.
 *
 * <p>Any modifications to this code may be overwritten and could lead to unexpected behavior.
 * Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public class Trace {

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer ct;
  private final MappedByteBuffer ctMax;
  private final MappedByteBuffer exceptionsAhoy;
  private final MappedByteBuffer first;
  private final MappedByteBuffer gasActual;
  private final MappedByteBuffer gasCost;
  private final MappedByteBuffer inputsAndOutputsAreMeaningful;
  private final MappedByteBuffer outOfGasException;
  private final MappedByteBuffer wcpArg1Lo;
  private final MappedByteBuffer wcpArg2Lo;
  private final MappedByteBuffer wcpInst;
  private final MappedByteBuffer wcpRes;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("gas.CT", 1, length));
      headers.add(new ColumnHeader("gas.CT_MAX", 1, length));
      headers.add(new ColumnHeader("gas.EXCEPTIONS_AHOY", 1, length));
      headers.add(new ColumnHeader("gas.FIRST", 1, length));
      headers.add(new ColumnHeader("gas.GAS_ACTUAL", 8, length));
      headers.add(new ColumnHeader("gas.GAS_COST", 8, length));
      headers.add(new ColumnHeader("gas.INPUTS_AND_OUTPUTS_ARE_MEANINGFUL", 1, length));
      headers.add(new ColumnHeader("gas.OUT_OF_GAS_EXCEPTION", 1, length));
      headers.add(new ColumnHeader("gas.WCP_ARG1_LO", 16, length));
      headers.add(new ColumnHeader("gas.WCP_ARG2_LO", 16, length));
      headers.add(new ColumnHeader("gas.WCP_INST", 1, length));
      headers.add(new ColumnHeader("gas.WCP_RES", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.ct = buffers.get(0);
    this.ctMax = buffers.get(1);
    this.exceptionsAhoy = buffers.get(2);
    this.first = buffers.get(3);
    this.gasActual = buffers.get(4);
    this.gasCost = buffers.get(5);
    this.inputsAndOutputsAreMeaningful = buffers.get(6);
    this.outOfGasException = buffers.get(7);
    this.wcpArg1Lo = buffers.get(8);
    this.wcpArg2Lo = buffers.get(9);
    this.wcpInst = buffers.get(10);
    this.wcpRes = buffers.get(11);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace ct(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("gas.CT already set");
    } else {
      filled.set(0);
    }

    if(b >= 8L) { throw new IllegalArgumentException("gas.CT has invalid value (" + b + ")"); }
    ct.put((byte) b);


    return this;
  }

  public Trace ctMax(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("gas.CT_MAX already set");
    } else {
      filled.set(1);
    }

    if(b >= 8L) { throw new IllegalArgumentException("gas.CT_MAX has invalid value (" + b + ")"); }
    ctMax.put((byte) b);


    return this;
  }

  public Trace exceptionsAhoy(final Boolean b) {
    if (filled.get(2)) {
      throw new IllegalStateException("gas.EXCEPTIONS_AHOY already set");
    } else {
      filled.set(2);
    }

    exceptionsAhoy.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace first(final Boolean b) {
    if (filled.get(3)) {
      throw new IllegalStateException("gas.FIRST already set");
    } else {
      filled.set(3);
    }

    first.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace gasActual(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("gas.GAS_ACTUAL already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("gas.GAS_ACTUAL has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasActual.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasActual.put(bs.get(j)); }

    return this;
  }

  public Trace gasCost(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("gas.GAS_COST already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("gas.GAS_COST has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasCost.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasCost.put(bs.get(j)); }

    return this;
  }

  public Trace inputsAndOutputsAreMeaningful(final Boolean b) {
    if (filled.get(6)) {
      throw new IllegalStateException("gas.INPUTS_AND_OUTPUTS_ARE_MEANINGFUL already set");
    } else {
      filled.set(6);
    }

    inputsAndOutputsAreMeaningful.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace outOfGasException(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("gas.OUT_OF_GAS_EXCEPTION already set");
    } else {
      filled.set(7);
    }

    outOfGasException.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace wcpArg1Lo(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("gas.WCP_ARG1_LO already set");
    } else {
      filled.set(8);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("gas.WCP_ARG1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { wcpArg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { wcpArg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace wcpArg2Lo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("gas.WCP_ARG2_LO already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("gas.WCP_ARG2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { wcpArg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { wcpArg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace wcpInst(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("gas.WCP_INST already set");
    } else {
      filled.set(10);
    }

    wcpInst.put(b.toByte());

    return this;
  }

  public Trace wcpRes(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("gas.WCP_RES already set");
    } else {
      filled.set(11);
    }

    wcpRes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("gas.CT has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("gas.CT_MAX has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("gas.EXCEPTIONS_AHOY has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("gas.FIRST has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("gas.GAS_ACTUAL has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("gas.GAS_COST has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("gas.INPUTS_AND_OUTPUTS_ARE_MEANINGFUL has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("gas.OUT_OF_GAS_EXCEPTION has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("gas.WCP_ARG1_LO has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("gas.WCP_ARG2_LO has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("gas.WCP_INST has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("gas.WCP_RES has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      ct.position(ct.position() + 1);
    }

    if (!filled.get(1)) {
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(2)) {
      exceptionsAhoy.position(exceptionsAhoy.position() + 1);
    }

    if (!filled.get(3)) {
      first.position(first.position() + 1);
    }

    if (!filled.get(4)) {
      gasActual.position(gasActual.position() + 8);
    }

    if (!filled.get(5)) {
      gasCost.position(gasCost.position() + 8);
    }

    if (!filled.get(6)) {
      inputsAndOutputsAreMeaningful.position(inputsAndOutputsAreMeaningful.position() + 1);
    }

    if (!filled.get(7)) {
      outOfGasException.position(outOfGasException.position() + 1);
    }

    if (!filled.get(8)) {
      wcpArg1Lo.position(wcpArg1Lo.position() + 16);
    }

    if (!filled.get(9)) {
      wcpArg2Lo.position(wcpArg2Lo.position() + 16);
    }

    if (!filled.get(10)) {
      wcpInst.position(wcpInst.position() + 1);
    }

    if (!filled.get(11)) {
      wcpRes.position(wcpRes.position() + 1);
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public void build() {
    if (!filled.isEmpty()) {
      throw new IllegalStateException("Cannot build trace with a non-validated row.");
    }
  }
}
