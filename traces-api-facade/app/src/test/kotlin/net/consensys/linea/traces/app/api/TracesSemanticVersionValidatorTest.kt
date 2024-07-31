package net.consensys.linea.traces.app.api

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequest
import net.consensys.linea.jsonrpc.JsonRpcRequestMapParams
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.MethodSource
import java.util.stream.Stream

class TracesSemanticVersionValidatorTest {
  private val validator = TracesSemanticVersionValidator(serverSemanticVersion)

  private fun buildRequestWithVersion(version: TracesSemanticVersionValidator.SemanticVersion): JsonRpcRequest {
    return JsonRpcRequestMapParams(
      "",
      1,
      "",
      mapOf(
        "block" to null,
        "rawExecutionTracesVersion" to null,
        "expectedTracesApiVersion" to version
      )
    )
  }

  @Test
  fun semanticVersion_isCreatedFromString_correctly() {
    val parsedSemanticVersion =
      TracesSemanticVersionValidator.SemanticVersion.fromString("1.2.3")
    assertEquals(
      TracesSemanticVersionValidator.SemanticVersion(1u, 2u, 3u),
      parsedSemanticVersion
    )
  }

  @ParameterizedTest
  @MethodSource("negativeTests")
  fun negativeTests(clientVersion: TracesSemanticVersionValidator.SemanticVersion) {
    val request = buildRequestWithVersion(clientVersion)
    assertEquals(
      Err(
        JsonRpcErrorResponse.invalidParams(
          request.id,
          "Client requested version $clientVersion is not compatible to server version $serverSemanticVersion"
        )
      ),
      validator.validateExpectedVersion(request.id, clientVersion.toString())
    )
  }

  @ParameterizedTest
  @MethodSource("positiveTests")
  fun positiveTests(clientVersion: TracesSemanticVersionValidator.SemanticVersion) {
    val request = buildRequestWithVersion(clientVersion)
    assertEquals(
      Ok(Unit),
      validator.validateExpectedVersion(request.id, clientVersion.toString())
    )
  }

  companion object {
    private val serverSemanticVersion = TracesSemanticVersionValidator.SemanticVersion(2u, 3u, 4u)

    @JvmStatic
    private fun negativeTests(): Stream<Arguments> {
      return Stream.of(
        Arguments.of(serverSemanticVersion.copy(major = 1u)),
        Arguments.of(serverSemanticVersion.copy(patch = 5u)),
        Arguments.of(serverSemanticVersion.copy(minor = 4u)),
        Arguments.of(serverSemanticVersion.copy(minor = 4u, patch = 7u)),
        Arguments.of(serverSemanticVersion.copy(major = 3u))
      )
    }

    @JvmStatic
    private fun positiveTests(): Stream<Arguments> {
      return Stream.of(
        Arguments.of(serverSemanticVersion),
        Arguments.of(serverSemanticVersion.copy(minor = 1u, patch = 0u)),
        Arguments.of(serverSemanticVersion.copy(minor = 1u)),
        Arguments.of(serverSemanticVersion.copy(minor = 2u, patch = 4u))
      )
    }
  }
}
