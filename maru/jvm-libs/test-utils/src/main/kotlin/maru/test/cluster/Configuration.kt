/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.test.cluster

import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.core.LoggerContext
import org.apache.logging.log4j.core.appender.ConsoleAppender
import org.apache.logging.log4j.core.config.builder.api.ConfigurationBuilderFactory

fun configureLoggers(
  pattern: String = "%d{yyyy-MM-dd HH:mm:ss} [%level] - %msg | %logger{36}%n",
  rootLevel: Level = Level.INFO,
  vararg logLevels: Pair<String, Level>,
) = configureLoggers(
  pattern,
  rootLevel,
  logLevels.toList(),
)

fun configureLoggers(
  pattern: String = "%d{yyyy-MM-dd HH:mm:ss} [%level] - %msg | %logger{36}%n",
  rootLevel: Level = Level.INFO,
  logLevels: List<Pair<String, Level>> = emptyList(),
) {
  val builder = ConfigurationBuilderFactory.newConfigurationBuilder()

  // Create pattern layout
  val layoutBuilder =
    builder
      .newLayout("PatternLayout")
      .addAttribute("pattern", pattern)

  // Create console appender
  val appenderBuilder =
    builder
      .newAppender("Console", "CONSOLE")
      .addAttribute("target", ConsoleAppender.Target.SYSTEM_OUT)
      .add(layoutBuilder)

  builder.add(appenderBuilder)

  // Configure root logger
  val rootLoggerBuilder =
    builder
      .newRootLogger(rootLevel)
      .add(builder.newAppenderRef("Console"))

  builder.add(rootLoggerBuilder)

  // Add specific logger configurations
  logLevels.forEach { (loggerName, level) ->
    builder.add(
      builder
        .newLogger(loggerName, level)
        .add(builder.newAppenderRef("Console"))
        .addAttribute("additivity", false),
    )
  }

  // Apply configuration
  val ctx = LogManager.getContext(false) as LoggerContext
  ctx.configuration = builder.build()
  ctx.updateLoggers()
}
