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
package tracing;

import java.security.InvalidParameterException;

public class TraceRequestParams {

  private static final int EXPECTED_PARAMS_SIZE = 3;

  private final long fromBlock;
  private final long toBlock;
  private final String runtimeVersion;

  public TraceRequestParams(final long fromBlock, final long toBlock, final String runtimeVersion) {
    this.fromBlock = fromBlock;
    this.toBlock = toBlock;
    this.runtimeVersion = runtimeVersion;
  }

  public long getFromBlock() {
    return fromBlock;
  }

  public long getToBlock() {
    return toBlock;
  }

  public String getRuntimeVersion() {
    return runtimeVersion;
  }

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
    return "1.0";
  }
}
