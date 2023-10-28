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

package net.consensys.linea.tracegeneration.rpc;

import java.security.InvalidParameterException;

import net.consensys.linea.zktracer.ZkTracer;

/** Holds needed parameters for sending an execution trace generation request. */
@SuppressWarnings("unused")
public record TraceRequestParams(long fromBlock, long toBlock, String runtimeVersion) {
  private static final int EXPECTED_PARAMS_SIZE = 3;

  /**
   * Parses a list of params to a {@link TraceRequestParams} object.
   *
   * @param params an array of parameters.
   * @return a parsed {@link TraceRequestParams} object..
   */
  public static TraceRequestParams createTraceParams(final Object[] params) {
    // validate params size
    if (params.length != EXPECTED_PARAMS_SIZE) {
      throw new InvalidParameterException(
          String.format("Expected %d parameters but got %d", EXPECTED_PARAMS_SIZE, params.length));
    }

    long fromBlock = Long.parseLong(params[0].toString());
    long toBlock = Long.parseLong(params[1].toString());
    String version = params[2].toString();

    if (!version.equals(getTracerRuntime())) {
      throw new InvalidParameterException(
          String.format(
              "INVALID_TRACES_VERSION: Runtime version is %s, requesting version %s",
              getTracerRuntime(), version));
    }

    return new TraceRequestParams(fromBlock, toBlock, version);
  }

  private static String getTracerRuntime() {
    return ZkTracer.class.getPackage().getSpecificationVersion();
  }
}
