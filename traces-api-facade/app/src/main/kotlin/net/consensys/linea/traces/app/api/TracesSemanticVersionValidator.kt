package net.consensys.linea.traces.app.api

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse

class TracesSemanticVersionValidator(private val apiVersion: SemanticVersion) {
  // Non-core portion is not supported
  data class SemanticVersion(val major: UInt, val minor: UInt, val patch: UInt) {
    companion object {
      private val simplifiedSemverRegex = """(\d+)\.(\d+)\.(\d+)""".toRegex()

      fun fromString(s: String): SemanticVersion {
        val groups = simplifiedSemverRegex.findAll(s).toList().first().groups
        return SemanticVersion(
          groups[1]!!.value.toUInt(),
          groups[2]!!.value.toUInt(),
          groups[3]!!.value.toUInt()
        )
      }
    }

    override fun toString(): String = "$major.$minor.$patch"

    // `this` considered to be server version and the argument to be client version
    fun isCompatible(clientRequestedVersion: SemanticVersion): Boolean {
      return when {
        (clientRequestedVersion.major != major) -> false
        (clientRequestedVersion.minor > minor) -> false
        (clientRequestedVersion.minor == minor && clientRequestedVersion.patch > patch) -> false
        else -> true
      }
    }
  }

  fun validateExpectedVersion(
    requestId: Any,
    expectedVersion: String
  ): Result<Unit, JsonRpcErrorResponse> {
    val clientRequestedVersion = SemanticVersion.fromString(expectedVersion)
    return when (apiVersion.isCompatible(clientRequestedVersion)) {
      true -> Ok(Unit)
      false -> Err(
        JsonRpcErrorResponse.invalidParams(
          requestId,
          "Client requested version $clientRequestedVersion is not compatible to server version $apiVersion"
        )
      )
    }
  }
}
