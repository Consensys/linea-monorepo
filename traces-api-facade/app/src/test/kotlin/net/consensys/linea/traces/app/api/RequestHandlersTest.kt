package net.consensys.linea.traces.app.api

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import net.consensys.linea.jsonrpc.JsonRpcErrorResponse
import net.consensys.linea.jsonrpc.JsonRpcRequestMapParams
import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test

class RequestHandlersTest {
  @Test
  fun validateParams_rejectsEmptyMap() {
    val request = JsonRpcRequestMapParams("", "", "", emptyMap<String, Any>())
    Assertions.assertEquals(
      Err(
        JsonRpcErrorResponse.invalidParams(
          request.id,
          "Parameters map is empty!"
        )
      ),
      validateParams(request)
    )
  }

  @Test
  fun validateParams_acceptsNonEmptyMap() {
    val request = JsonRpcRequestMapParams("", "", "", mapOf("key" to "value"))
    Assertions.assertEquals(Ok(request), validateParams(request))
  }
}
