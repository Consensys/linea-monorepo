package net.consensys.linea.logging

import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.logging.log4j.core.appender.ConsoleAppender
import org.apache.logging.log4j.core.appender.OutputStreamAppender
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.Test
import java.io.ByteArrayOutputStream

class DebouncingFilterIntegrationTest {
  @Test
  fun `DebouncingFilter is configurable with resources`() {
    // Configuration is defined in file log4j2.xml in test resources
    val log: Logger = LogManager.getLogger(DebouncingFilterIntegrationTest::class.java)
    val coreLogger = (log as org.apache.logging.log4j.core.Logger)
    val outContent = ByteArrayOutputStream()
    val appender = OutputStreamAppender.newBuilder()
      .setName("Test appender")
      .setTarget(outContent)
      .setFilter((coreLogger.context.configuration.appenders.values.first() as ConsoleAppender).filter)
      .build()

    coreLogger.addAppender(appender)
    val loggedString = "AAAAAA!"
    for (i in 1..10) {
      log.info(loggedString)
    }

    Assertions.assertThat(countNumberOfEntries(outContent.toString(), loggedString)).isEqualTo(1)
  }

  private fun countNumberOfEntries(stringToSearchIn: String, stringToSearchFor: String): Int {
    return Regex.fromLiteral(stringToSearchFor).findAll(stringToSearchIn).toList().size
  }
}
