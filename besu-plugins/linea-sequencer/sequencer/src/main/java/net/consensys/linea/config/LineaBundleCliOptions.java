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

package net.consensys.linea.config;

import java.net.URL;
import java.util.Set;
import java.util.stream.Collectors;

import com.google.common.base.MoreObjects;
import net.consensys.linea.plugins.LineaCliOptions;
import picocli.CommandLine;

/** The Linea Bundle CLI options. */
public class LineaBundleCliOptions implements LineaCliOptions {
  public static final String CONFIG_KEY = "bundle-sequencer";

  private static final String BUNDLES_FORWARD_URLS = "--plugin-linea-bundles-forward-urls";
  private static final Set<URL> DEFAULT_BUNDLES_FORWARD_URLS = Set.of();

  @CommandLine.Option(
      names = {BUNDLES_FORWARD_URLS},
      paramLabel = "<SET<URL>>",
      description =
          "A comma separated list of endpoint to which incoming bundles will be forwarded (default: ${DEFAULT-VALUE})")
  private Set<URL> bundleForwardUrls = DEFAULT_BUNDLES_FORWARD_URLS;

  private LineaBundleCliOptions() {}

  /**
   * Create Linea RPC CLI options.
   *
   * @return the Linea RPC CLI options
   */
  public static LineaBundleCliOptions create() {
    return new LineaBundleCliOptions();
  }

  /**
   * Linea RPC CLI options from config.
   *
   * @param config the config
   * @return the Linea RPC CLI options
   */
  public static LineaBundleCliOptions fromConfig(final LineaBundleConfiguration config) {
    final LineaBundleCliOptions options = create();
    options.bundleForwardUrls = config.bundleForwardUrls();
    return options;
  }

  /**
   * To domain object Linea factory configuration.
   *
   * @return the Linea factory configuration
   */
  @Override
  public LineaBundleConfiguration toDomainObject() {
    return LineaBundleConfiguration.builder().bundleForwardUrls(bundleForwardUrls).build();
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(
            BUNDLES_FORWARD_URLS,
            bundleForwardUrls.stream().map(URL::toString).collect(Collectors.joining(",")))
        .toString();
  }
}
