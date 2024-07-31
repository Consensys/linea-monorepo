package net.consensys.linea.jsonrpc

import com.fasterxml.jackson.annotation.JsonIgnore
import com.fasterxml.jackson.annotation.JsonProperty
import java.util.StringJoiner

interface JsonRpcRequest {
  val jsonrpc: String
  val id: Any
  val method: String
  val params: Any

  @get:JsonIgnore
  val isValid: Boolean
    get() {
      if ("2.0" != jsonrpc) {
        return false
      }
      return "" != method
    }
}

data class JsonRpcRequestListParams(
  @JsonProperty("jsonrpc") override val jsonrpc: String,
  @JsonProperty("id") override val id: Any,
  @JsonProperty("method") override val method: String,
  @JsonProperty("params") override val params: List<Any?>
) : JsonRpcRequest {
  override fun toString(): String {
    return StringJoiner(", ", JsonRpcRequestListParams::class.java.simpleName + "[", "]")
      .add("jsonrpc='$jsonrpc'")
      .add("id='$id'")
      .add("method='$method'")
      .add("params=$params")
      .toString()
  }
}

data class JsonRpcRequestMapParams(
  @JsonProperty("jsonrpc") override val jsonrpc: String,
  @JsonProperty("id") override val id: Any,
  @JsonProperty("method") override val method: String,
  @JsonProperty("params") override val params: Map<String, *>
) : JsonRpcRequest {
  override fun toString(): String {
    return StringJoiner(", ", JsonRpcRequestMapParams::class.java.simpleName + "[", "]")
      .add("jsonrpc='$jsonrpc'")
      .add("id='$id'")
      .add("method='$method'")
      .add("params=$params")
      .toString()
  }
}
