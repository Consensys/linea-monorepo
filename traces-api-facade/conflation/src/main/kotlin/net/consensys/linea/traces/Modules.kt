package net.consensys.linea.traces

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty
import io.vertx.core.json.JsonObject
import kotlin.Pair

fun JsonObject.getTrace(jsonPath: List<String>): JsonObject? {
  var jsonObject: JsonObject? = this
  for (node in jsonPath) {
    jsonObject = jsonObject?.getJsonObject(node)
  }
  return jsonObject?.getJsonObject("Trace")
}

val MODULES = listOf(
  Pair(listOf("add"), Add::class.java),
  Pair(listOf("bin"), Bin::class.java),
  Pair(listOf("binRT"), BinRt::class.java),
  Pair(listOf("ec_data"), EcData::class.java),
  Pair(listOf("ext"), Ext::class.java),
  Pair(listOf("pub", "hash_data"), HashData::class.java),
  Pair(listOf("pub", "hash_info"), HashInfo::class.java),
  Pair(listOf("instruction-decoder"), InstructionDecoder::class.java),
  Pair(listOf("pub", "log_data"), LogData::class.java),
  Pair(listOf("pub", "log_info"), LogInfo::class.java),
  Pair(listOf("hub", "mmu", "mmio"), Mmio::class.java),
  Pair(listOf("hub", "mmu"), Mmu::class.java),
  Pair(listOf("mmuID"), MmuId::class.java),
  Pair(listOf("mod"), Mod::class.java),
  Pair(listOf("mul"), Mul::class.java),
  Pair(listOf("mxp"), Mxp::class.java),
  Pair(listOf("phoneyRLP"), PhoneyRlp::class.java),
  Pair(listOf("rlp"), Rlp::class.java),
  Pair(listOf("rom"), Rom::class.java),
  Pair(listOf("shf"), Shf::class.java),
  Pair(listOf("shfRT"), ShfRt::class.java),
  Pair(listOf("txRlp"), TxRlp::class.java),
  Pair(listOf("wcp"), Wcp::class.java),
  Pair(listOf("hub"), Hub::class.java)
)

interface EVMTracesState {
  fun cleanState() {
    for (field in this.javaClass.declaredFields) {
      field.isAccessible = true
      if (field.type == MutableList::class.java) {
        field.set(this, mutableListOf<String>())
      }
    }
  }
}

@JsonIgnoreProperties(ignoreUnknown = true)
data class Hub(
  @get:JsonProperty("ALPHA")
  val ALPHA: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALU_ADD_INST")
  val ALU_ADD_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALU_EXT_INST")
  val ALU_EXT_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALU_MOD_INST")
  val ALU_MOD_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALU_MUL_INST")
  val ALU_MUL_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALU_STAMP")
  val ALU_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARITHMETIC_INST")
  val ARITHMETIC_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BINARY_INST")
  val BINARY_INST: MutableList<String> = arrayListOf(),
  // @get:JsonProperty("BIN_STAMP")
  // val BIN_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTECODE_ADDRESS_HI")
  val BYTECODE_ADDRESS_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTECODE_ADDRESS_LO")
  val BYTECODE_ADDRESS_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CALLDATA_OFFSET")
  val CALLDATA_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CALLDATA_SIZE")
  val CALLDATA_SIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CALLER_CONTEXT")
  val CALLER_CONTEXT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CALLSTACK_DEPTH")
  val CALLSTACK_DEPTH: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CODECALL_FLAG")
  val CODECALL_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_ERROR")
  val CONTEXT_ERROR: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_NUMBER")
  val CONTEXT_NUMBER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_REVERTS")
  val CONTEXT_REVERTS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_REVERTS_BY_CHOICE")
  val CONTEXT_REVERTS_BY_CHOICE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_REVERTS_BY_FORCE")
  val CONTEXT_REVERTS_BY_FORCE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_REVERT_STORAGE_STAMP")
  val CONTEXT_REVERT_STORAGE_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_RUNS_OUT_OF_GAS")
  val CONTEXT_RUNS_OUT_OF_GAS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_TYPE")
  val CONTEXT_TYPE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COUNTER")
  val COUNTER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("DELEGATECALL_FLAG")
  val DELEGATECALL_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("DELTA")
  val DELTA: MutableList<String> = arrayListOf(),
  @get:JsonProperty("FLAG_1")
  val FLAG_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("FLAG_2")
  val FLAG_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("FLAG_3")
  val FLAG_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("HEIGHT")
  val HEIGHT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("HEIGHT_NEW")
  val HEIGHT_NEW: MutableList<String> = arrayListOf(),
  @get:JsonProperty("HEIGHT_OVER")
  val HEIGHT_OVER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("HEIGHT_UNDER")
  val HEIGHT_UNDER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INSTRUCTION")
  val INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INSTRUCTION_ARGUMENT_HI")
  val INSTRUCTION_ARGUMENT_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INSTRUCTION_ARGUMENT_LO")
  val INSTRUCTION_ARGUMENT_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INSTRUCTION_STAMP")
  val INSTRUCTION_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INST_PARAM")
  val INST_PARAM: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_INITCODE")
  val IS_INITCODE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ITEM_HEIGHT_1")
  val ITEM_HEIGHT_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ITEM_HEIGHT_2")
  val ITEM_HEIGHT_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ITEM_HEIGHT_3")
  val ITEM_HEIGHT_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ITEM_HEIGHT_4")
  val ITEM_HEIGHT_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ITEM_STACK_STAMP_1")
  val ITEM_STACK_STAMP_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ITEM_STACK_STAMP_2")
  val ITEM_STACK_STAMP_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ITEM_STACK_STAMP_3")
  val ITEM_STACK_STAMP_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ITEM_STACK_STAMP_4")
  val ITEM_STACK_STAMP_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MAXIMUM_CONTEXT")
  val MAXIMUM_CONTEXT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PC")
  val PC: MutableList<String> = arrayListOf(),
  @get:JsonProperty("POP_1")
  val POP_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("POP_2")
  val POP_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("POP_3")
  val POP_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("POP_4")
  val POP_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RAM_INST")
  val RAM_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RAM_STAMP")
  val RAM_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RETURNDATA_OFFSET")
  val RETURNDATA_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RETURNDATA_SIZE")
  val RETURNDATA_SIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RETURNER_CONTEXT")
  val RETURNER_CONTEXT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RETURN_CAPACITY")
  val RETURN_CAPACITY: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RETURN_OFFSET")
  val RETURN_OFFSET: MutableList<String> = arrayListOf(),
  // @get:JsonProperty("SHF_STAMP")
  // val SHF_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHIFT_INST")
  val SHIFT_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STACK_EXCEPTION")
  val STACK_EXCEPTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STACK_OVERFLOW_EXCEPTION")
  val STACK_OVERFLOW_EXCEPTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STACK_PATTERN")
  val STACK_PATTERN: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STACK_STAMP")
  val STACK_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STACK_STAMP_NEW")
  val STACK_STAMP_NEW: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STACK_UNDERFLOW_EXCEPTION")
  val STACK_UNDERFLOW_EXCEPTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STATICCALL_FLAG")
  val STATICCALL_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STATIC_GAS")
  val STATIC_GAS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STORAGE_INST")
  val STORAGE_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STO_STAMP")
  val STO_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TWO_LINES_INSTRUCTION")
  val TWO_LINES_INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TX_NUM")
  val TX_NUM: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VALUE")
  val VALUE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_HI_1")
  val VAL_HI_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_HI_2")
  val VAL_HI_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_HI_3")
  val VAL_HI_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_HI_4")
  val VAL_HI_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_LO_1")
  val VAL_LO_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_LO_2")
  val VAL_LO_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_LO_3")
  val VAL_LO_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_LO_4")
  val VAL_LO_4: MutableList<String> = arrayListOf(),
  // @get:JsonProperty("WCP_STAMP")
  // val WCP_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WORD_COMPARISON_INST")
  val WORD_COMPARISON_INST: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Add(
  @get:JsonProperty("ACC_1")
  val ACC_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_2")
  val ACC_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_1_HI")
  val ARG_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_1_LO")
  val ARG_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_2_HI")
  val ARG_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_2_LO")
  val ARG_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_1")
  val BYTE_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_2")
  val BYTE_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CT")
  val CT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INST")
  val INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OVERFLOW")
  val OVERFLOW: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RES_HI")
  val RES_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RES_LO")
  val RES_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STAMP")
  val STAMP: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Bin(
  @get:JsonProperty("ACC_1")
  val ACC_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_2")
  val ACC_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_3")
  val ACC_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_4")
  val ACC_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_5")
  val ACC_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_6")
  val ACC_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("AND_BYTE_HI")
  val AND_BYTE_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("AND_BYTE_LO")
  val AND_BYTE_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARGUMENT_1_HI")
  val ARGUMENT_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARGUMENT_1_LO")
  val ARGUMENT_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARGUMENT_2_HI")
  val ARGUMENT_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARGUMENT_2_LO")
  val ARGUMENT_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BINARY_STAMP")
  val BINARY_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BITS")
  val BITS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_1")
  val BIT_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_B_4")
  val BIT_B_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_1")
  val BYTE_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_2")
  val BYTE_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_3")
  val BYTE_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_4")
  val BYTE_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_5")
  val BYTE_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_6")
  val BYTE_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COUNTER")
  val COUNTER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INST")
  val INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_DATA")
  val IS_DATA: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LOW_4")
  val LOW_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NEG")
  val NEG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NOT_BYTE_HI")
  val NOT_BYTE_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NOT_BYTE_LO")
  val NOT_BYTE_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ONE_LINE_INSTRUCTION")
  val ONE_LINE_INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OR_BYTE_HI")
  val OR_BYTE_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OR_BYTE_LO")
  val OR_BYTE_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PIVOT")
  val PIVOT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RESULT_HI")
  val RESULT_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RESULT_LO")
  val RESULT_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SMALL")
  val SMALL: MutableList<String> = arrayListOf(),
  @get:JsonProperty("XOR_BYTE_HI")
  val XOR_BYTE_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("XOR_BYTE_LO")
  val XOR_BYTE_LO: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class BinRt(
  @get:JsonProperty("AND_BYTE")
  val AND_BYTE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_ARG_1")
  val BYTE_ARG_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_ARG_2")
  val BYTE_ARG_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_IN_RT")
  val IS_IN_RT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NOT_BYTE")
  val NOT_BYTE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OR_BYTE")
  val OR_BYTE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("XOR_BYTE")
  val XOR_BYTE: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class EcData(
  @get:JsonProperty("ACC_DELTA")
  val ACC_DELTA: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_PAIRINGS")
  val ACC_PAIRINGS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALL_CHECKS_PASSED")
  val ALL_CHECKS_PASSED: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_DELTA")
  val BYTE_DELTA: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COMPARISONS")
  val COMPARISONS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CT_MIN")
  val CT_MIN: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CUBE")
  val CUBE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EC_ADD")
  val EC_ADD: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EC_MUL")
  val EC_MUL: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EC_PAIRING")
  val EC_PAIRING: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EC_RECOVER")
  val EC_RECOVER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EQUALITIES")
  val EQUALITIES: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXT_ARG1_HI")
  val EXT_ARG_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXT_ARG1_LO")
  val EXT_ARG_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXT_ARG2_HI")
  val EXT_ARG_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXT_ARG2_LO")
  val EXT_ARG_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXT_ARG3_HI")
  val EXT_ARG_3_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXT_ARG3_LO")
  val EXT_ARG_3_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXT_INST")
  val EXT_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXT_RES_HI")
  val EXT_RES_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXT_RES_LO")
  val EXT_RES_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("HURDLE")
  val HURDLE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INDEX")
  val INDEX: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LIMB")
  val LIMB: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PRELIMINARY_CHECKS_PASSED")
  val PRELIMINARY_CHECKS_PASSED: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SOMETHING_WASNT_ON_G2")
  val SOMETHING_WASNT_ON_G_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SQUARE")
  val SQUARE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STAMP")
  val STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("THIS_IS_NOT_ON_G2")
  val THIS_IS_NOT_ON_G_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("THIS_IS_NOT_ON_G2_ACC")
  val THIS_IS_NOT_ON_G_2_ACC: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOTAL_PAIRINGS")
  val TOTAL_PAIRINGS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TYPE")
  val TYPE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WCP_ARG1_HI")
  val WCP_ARG_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WCP_ARG1_LO")
  val WCP_ARG_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WCP_ARG2_HI")
  val WCP_ARG_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WCP_ARG2_LO")
  val WCP_ARG_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WCP_INST")
  val WCP_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WCP_RES")
  val WCP_RES: MutableList<String> = arrayListOf()
) : EVMTracesState

@JsonIgnoreProperties(ignoreUnknown = true)
data class Ext(
  @get:JsonProperty("ACC_A_0")
  val ACC_A_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_A_1")
  val ACC_A_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_A_2")
  val ACC_A_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_A_3")
  val ACC_A_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_0")
  val ACC_B_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_1")
  val ACC_B_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_2")
  val ACC_B_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_3")
  val ACC_B_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_C_0")
  val ACC_C_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_C_1")
  val ACC_C_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_C_2")
  val ACC_C_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_C_3")
  val ACC_C_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_DELTA_0")
  val ACC_DELTA_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_DELTA_1")
  val ACC_DELTA_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_DELTA_2")
  val ACC_DELTA_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_DELTA_3")
  val ACC_DELTA_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_0")
  val ACC_H_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_1")
  val ACC_H_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_2")
  val ACC_H_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_3")
  val ACC_H_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_4")
  val ACC_H_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_5")
  val ACC_H_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_I_0")
  val ACC_I_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_I_1")
  val ACC_I_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_I_2")
  val ACC_I_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_I_3")
  val ACC_I_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_I_4")
  val ACC_I_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_I_5")
  val ACC_I_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_I_6")
  val ACC_I_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_J_0")
  val ACC_J_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_J_1")
  val ACC_J_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_J_2")
  val ACC_J_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_J_3")
  val ACC_J_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_J_4")
  val ACC_J_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_J_5")
  val ACC_J_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_J_6")
  val ACC_J_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_J_7")
  val ACC_J_7: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_0")
  val ACC_Q_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_1")
  val ACC_Q_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_2")
  val ACC_Q_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_3")
  val ACC_Q_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_4")
  val ACC_Q_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_5")
  val ACC_Q_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_6")
  val ACC_Q_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_7")
  val ACC_Q_7: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_R_0")
  val ACC_R_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_R_1")
  val ACC_R_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_R_2")
  val ACC_R_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_R_3")
  val ACC_R_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_1_HI")
  val ARG_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_1_LO")
  val ARG_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_2_HI")
  val ARG_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_2_LO")
  val ARG_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_3_HI")
  val ARG_3_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_3_LO")
  val ARG_3_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_1")
  val BIT_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_2")
  val BIT_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_3")
  val BIT_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_A_0")
  val BYTE_A_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_A_1")
  val BYTE_A_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_A_2")
  val BYTE_A_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_A_3")
  val BYTE_A_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_0")
  val BYTE_B_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_1")
  val BYTE_B_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_2")
  val BYTE_B_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_3")
  val BYTE_B_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_C_0")
  val BYTE_C_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_C_1")
  val BYTE_C_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_C_2")
  val BYTE_C_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_C_3")
  val BYTE_C_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_DELTA_0")
  val BYTE_DELTA_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_DELTA_1")
  val BYTE_DELTA_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_DELTA_2")
  val BYTE_DELTA_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_DELTA_3")
  val BYTE_DELTA_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_0")
  val BYTE_H_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_1")
  val BYTE_H_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_2")
  val BYTE_H_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_3")
  val BYTE_H_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_4")
  val BYTE_H_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_5")
  val BYTE_H_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_I_0")
  val BYTE_I_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_I_1")
  val BYTE_I_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_I_2")
  val BYTE_I_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_I_3")
  val BYTE_I_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_I_4")
  val BYTE_I_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_I_5")
  val BYTE_I_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_I_6")
  val BYTE_I_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_J_0")
  val BYTE_J_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_J_1")
  val BYTE_J_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_J_2")
  val BYTE_J_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_J_3")
  val BYTE_J_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_J_4")
  val BYTE_J_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_J_5")
  val BYTE_J_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_J_6")
  val BYTE_J_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_J_7")
  val BYTE_J_7: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_0")
  val BYTE_Q_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_1")
  val BYTE_Q_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_2")
  val BYTE_Q_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_3")
  val BYTE_Q_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_4")
  val BYTE_Q_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_5")
  val BYTE_Q_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_6")
  val BYTE_Q_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_7")
  val BYTE_Q_7: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_R_0")
  val BYTE_R_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_R_1")
  val BYTE_R_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_R_2")
  val BYTE_R_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_R_3")
  val BYTE_R_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CMP")
  val CMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CT")
  val CT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INST")
  val INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OF_H")
  val OF_H: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OF_I")
  val OF_I: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OF_J")
  val OF_J: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OF_RES")
  val OF_RES: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OLI")
  val OLI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RES_HI")
  val RES_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RES_LO")
  val RES_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STAMP")
  val STAMP: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class HashData(
  @get:JsonProperty("INDEX")
  val INDEX: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INST")
  val INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LIMB")
  val LIMB: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NUM")
  val NUM: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class HashInfo(
  @get:JsonProperty("HASH_HI")
  val HASH_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("HASH_LO")
  val HASH_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("HASH_NUM")
  val HASH_NUM: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class InstructionDecoder(
  @get:JsonProperty("ADD_FLAG")
  val ADD_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ADD_MOD_FLAG")
  val ADD_MOD_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALPHA")
  val ALPHA: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALU_ADD_INST")
  val ALU_ADD_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALU_EXT_INST")
  val ALU_EXT_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALU_MOD_INST")
  val ALU_MOD_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALU_MUL_INST")
  val ALU_MUL_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARITHMETIC_INST")
  val ARITHMETIC_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BINARY_INST")
  val BINARY_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CALLDATACOPY_FLAG")
  val CALLDATACOPY_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CALLDATA_FLAG")
  val CALLDATA_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CALL_FLAG")
  val CALL_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("DELTA")
  val DELTA: MutableList<String> = arrayListOf(),
  @get:JsonProperty("DIV_FLAG")
  val DIV_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXODATA_IS_SOURCE")
  val EXODATA_IS_SOURCE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXOOP_FLAG")
  val EXOOP_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXP_FLAG")
  val EXP_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("FLAG_1")
  val FLAG_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("FLAG_2")
  val FLAG_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("FLAG_3")
  val FLAG_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("HASH_INST")
  val HASH_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INST")
  val INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INST_PARAM")
  val INST_PARAM: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INVALID_INSTRUCTION")
  val INVALID_INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("JUMP_FLAG")
  val JUMP_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LOG_INST")
  val LOG_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MEMORY_EXPANSION_FLAG")
  val MEMORY_EXPANSION_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MOD_FLAG")
  val MOD_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MUL_FLAG")
  val MUL_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MUL_MOD_FLAG")
  val MUL_MOD_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_GBYTE")
  val MXP_GBYTE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_GWORD")
  val MXP_GWORD: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_INST")
  val MXP_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_TYPE_1")
  val MXP_TYPE_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_TYPE_2")
  val MXP_TYPE_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_TYPE_3")
  val MXP_TYPE_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_TYPE_4")
  val MXP_TYPE_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_TYPE_5")
  val MXP_TYPE_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NB_ADDED")
  val NB_ADDED: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NB_REMOVED")
  val NB_REMOVED: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NON_STATIC_FLAG")
  val NON_STATIC_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OP")
  val OP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PUSH_FLAG")
  val PUSH_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RAM_INST")
  val RAM_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RETURNDATA_FLAG")
  val RETURNDATA_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RETURN_FLAG")
  val RETURN_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("REVERT_FLAG")
  val REVERT_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ROM_FLAG")
  val ROM_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SDIV_FLAG")
  val SDIV_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHIFT_INST")
  val SHIFT_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SIZE")
  val SIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SMOD_FLAG")
  val SMOD_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SPECIAL_PC_UPDATE")
  val SPECIAL_PC_UPDATE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STACK_PATTERN")
  val STACK_PATTERN: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STATIC_GAS")
  val STATIC_GAS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STOP_FLAG")
  val STOP_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STORAGE_INST")
  val STORAGE_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SUB_FLAG")
  val SUB_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TWO_LINES_INSTRUCTION")
  val TWO_LINES_INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WARMTH_FLAG")
  val WARMTH_FLAG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WORD_COMPARISON_INST")
  val WORD_COMPARISON_INST: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class LogData(
  @get:JsonProperty("INDEX")
  val INDEX: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LIMB")
  val LIMB: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NUM")
  val NUM: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class LogInfo(
  @get:JsonProperty("ADDR_HI")
  val ADDR_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ADDR_LO")
  val ADDR_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LOG_NUM")
  val LOG_NUM: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OP")
  val OP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SIZE")
  val SIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOPIC_0_HI")
  val TOPIC_0_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOPIC_0_LO")
  val TOPIC_0_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOPIC_1_HI")
  val TOPIC_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOPIC_1_LO")
  val TOPIC_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOPIC_2_HI")
  val TOPIC_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOPIC_2_LO")
  val TOPIC_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOPIC_3_HI")
  val TOPIC_3_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOPIC_3_LO")
  val TOPIC_3_LO: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Mmio(
  @get:JsonProperty("ACC_1")
  val ACC_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_2")
  val ACC_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_3")
  val ACC_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_4")
  val ACC_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_5")
  val ACC_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_6")
  val ACC_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_A")
  val ACC_A: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B")
  val ACC_B: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_C")
  val ACC_C: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_VAL_HI")
  val ACC_VAL_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_VAL_LO")
  val ACC_VAL_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_X")
  val ACC_X: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIN_1")
  val BIN_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIN_2")
  val BIN_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIN_3")
  val BIN_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIN_4")
  val BIN_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIN_5")
  val BIN_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_A")
  val BYTE_A: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B")
  val BYTE_B: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_C")
  val BYTE_C: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_X")
  val BYTE_X: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CN_A")
  val CN_A: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CN_B")
  val CN_B: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CN_C")
  val CN_C: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_SOURCE")
  val CONTEXT_SOURCE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_TARGET")
  val CONTEXT_TARGET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COUNTER")
  val COUNTER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ERF")
  val ERF: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXO_IS_HASH")
  val EXO_IS_HASH: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXO_IS_LOG")
  val EXO_IS_LOG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXO_IS_ROM")
  val EXO_IS_ROM: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXO_IS_TXCD")
  val EXO_IS_TXCD: MutableList<String> = arrayListOf(),
  @get:JsonProperty("FAST")
  val FAST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INDEX_A")
  val INDEX_A: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INDEX_B")
  val INDEX_B: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INDEX_C")
  val INDEX_C: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INDEX_X")
  val INDEX_X: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_INIT")
  val IS_INIT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LOG_NUM")
  val LOG_NUM: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MICRO_INSTRUCTION")
  val MICRO_INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MICRO_INSTRUCTION_STAMP")
  val MICRO_INSTRUCTION_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("POW_256_1")
  val POW_256_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("POW_256_2")
  val POW_256_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SIZE")
  val SIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SOURCE_BYTE_OFFSET")
  val SOURCE_BYTE_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SOURCE_LIMB_OFFSET")
  val SOURCE_LIMB_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STACK_VALUE_HIGH")
  val STACK_VALUE_HIGH: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STACK_VALUE_HI_BYTE")
  val STACK_VALUE_HI_BYTE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STACK_VALUE_LOW")
  val STACK_VALUE_LOW: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STACK_VALUE_LO_BYTE")
  val STACK_VALUE_LO_BYTE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TARGET_BYTE_OFFSET")
  val TARGET_BYTE_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TARGET_LIMB_OFFSET")
  val TARGET_LIMB_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TX_NUM")
  val TX_NUM: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_A")
  val VAL_A: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_A_NEW")
  val VAL_A_NEW: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_B")
  val VAL_B: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_B_NEW")
  val VAL_B_NEW: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_C")
  val VAL_C: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_C_NEW")
  val VAL_C_NEW: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_X")
  val VAL_X: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Mmu(
  @get:JsonProperty("ACC_1")
  val ACC_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_2")
  val ACC_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_3")
  val ACC_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_4")
  val ACC_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_5")
  val ACC_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_6")
  val ACC_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_7")
  val ACC_7: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_8")
  val ACC_8: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ALIGNED")
  val ALIGNED: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_1")
  val BIT_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_2")
  val BIT_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_3")
  val BIT_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_4")
  val BIT_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_5")
  val BIT_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_6")
  val BIT_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_7")
  val BIT_7: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_8")
  val BIT_8: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_1")
  val BYTE_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_2")
  val BYTE_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_3")
  val BYTE_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_4")
  val BYTE_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_5")
  val BYTE_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_6")
  val BYTE_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_7")
  val BYTE_7: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_8")
  val BYTE_8: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CALLER")
  val CALLER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CALL_DATA_OFFSET")
  val CALL_DATA_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CALL_DATA_SIZE")
  val CALL_DATA_SIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CALL_STACK_DEPTH")
  val CALL_STACK_DEPTH: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_NUMBER")
  val CONTEXT_NUMBER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_SOURCE")
  val CONTEXT_SOURCE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CONTEXT_TARGET")
  val CONTEXT_TARGET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COUNTER")
  val COUNTER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ERF")
  val ERF: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXO_IS_HASH")
  val EXO_IS_HASH: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXO_IS_LOG")
  val EXO_IS_LOG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXO_IS_ROM")
  val EXO_IS_ROM: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXO_IS_TXCD")
  val EXO_IS_TXCD: MutableList<String> = arrayListOf(),
  @get:JsonProperty("FAST")
  val FAST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INFO")
  val INFO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INSTRUCTION")
  val INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_DATA")
  val IS_DATA: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_MICRO_INSTRUCTION")
  val IS_MICRO_INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MICRO_INSTRUCTION")
  val MICRO_INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MICRO_INSTRUCTION_STAMP")
  val MICRO_INSTRUCTION_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MIN")
  val MIN: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NIB_1")
  val NIB_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NIB_2")
  val NIB_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NIB_3")
  val NIB_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NIB_4")
  val NIB_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NIB_5")
  val NIB_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NIB_6")
  val NIB_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NIB_7")
  val NIB_7: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NIB_8")
  val NIB_8: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NIB_9")
  val NIB_9: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OFFSET_OUT_OF_BOUNDS")
  val OFFSET_OUT_OF_BOUNDS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OFF_1_LO")
  val OFF_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OFF_2_HI")
  val OFF_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OFF_2_LO")
  val OFF_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PRECOMPUTATION")
  val PRECOMPUTATION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RAM_STAMP")
  val RAM_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("REFO")
  val REFO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("REFS")
  val REFS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RETURNER")
  val RETURNER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RETURN_CAPACITY")
  val RETURN_CAPACITY: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RETURN_OFFSET")
  val RETURN_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SIZE")
  val SIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SIZE_IMPORTED")
  val SIZE_IMPORTED: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SOURCE_BYTE_OFFSET")
  val SOURCE_BYTE_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SOURCE_LIMB_OFFSET")
  val SOURCE_LIMB_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TARGET_BYTE_OFFSET")
  val TARGET_BYTE_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TARGET_LIMB_OFFSET")
  val TARGET_LIMB_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TERNARY")
  val TERNARY: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOTAL_NUMBER_OF_MICRO_INSTRUCTIONS")
  val TOTAL_NUMBER_OF_MICRO_INSTRUCTIONS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOTAL_NUMBER_OF_PADDINGS")
  val TOTAL_NUMBER_OF_PADDINGS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TOTAL_NUMBER_OF_READS")
  val TOTAL_NUMBER_OF_READS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TO_RAM")
  val TO_RAM: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_HI")
  val VAL_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("VAL_LO")
  val VAL_LO: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class MmuId(
  @get:JsonProperty("INFO")
  val INFO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INST")
  val INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_IN_ID")
  val IS_IN_ID: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PRE")
  val PRE: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Mod(
  @get:JsonProperty("ACC_1_2")
  val ACC_1_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_1_3")
  val ACC_1_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_2_2")
  val ACC_2_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_2_3")
  val ACC_2_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_0")
  val ACC_B_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_1")
  val ACC_B_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_2")
  val ACC_B_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_3")
  val ACC_B_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_DELTA_0")
  val ACC_DELTA_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_DELTA_1")
  val ACC_DELTA_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_DELTA_2")
  val ACC_DELTA_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_DELTA_3")
  val ACC_DELTA_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_0")
  val ACC_H_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_1")
  val ACC_H_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_2")
  val ACC_H_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_0")
  val ACC_Q_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_1")
  val ACC_Q_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_2")
  val ACC_Q_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q_3")
  val ACC_Q_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_R_0")
  val ACC_R_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_R_1")
  val ACC_R_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_R_2")
  val ACC_R_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_R_3")
  val ACC_R_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_1_HI")
  val ARG_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_1_LO")
  val ARG_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_2_HI")
  val ARG_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_2_LO")
  val ARG_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_1_2")
  val BYTE_1_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_1_3")
  val BYTE_1_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_2_2")
  val BYTE_2_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_2_3")
  val BYTE_2_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_0")
  val BYTE_B_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_1")
  val BYTE_B_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_2")
  val BYTE_B_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_3")
  val BYTE_B_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_DELTA_0")
  val BYTE_DELTA_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_DELTA_1")
  val BYTE_DELTA_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_DELTA_2")
  val BYTE_DELTA_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_DELTA_3")
  val BYTE_DELTA_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_0")
  val BYTE_H_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_1")
  val BYTE_H_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_2")
  val BYTE_H_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_0")
  val BYTE_Q_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_1")
  val BYTE_Q_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_2")
  val BYTE_Q_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q_3")
  val BYTE_Q_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_R_0")
  val BYTE_R_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_R_1")
  val BYTE_R_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_R_2")
  val BYTE_R_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_R_3")
  val BYTE_R_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CMP_1")
  val CMP_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CMP_2")
  val CMP_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CT")
  val CT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("DEC_OUTPUT")
  val DEC_OUTPUT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("DEC_SIGNED")
  val DEC_SIGNED: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INST")
  val INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MSB_1")
  val MSB_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MSB_2")
  val MSB_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OLI")
  val OLI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RES_HI")
  val RES_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RES_LO")
  val RES_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STAMP")
  val STAMP: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Mul(
  @get:JsonProperty("ACC_A_0")
  val ACC_A_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_A_1")
  val ACC_A_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_A_2")
  val ACC_A_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_A_3")
  val ACC_A_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_0")
  val ACC_B_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_1")
  val ACC_B_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_2")
  val ACC_B_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_B_3")
  val ACC_B_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_C_0")
  val ACC_C_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_C_1")
  val ACC_C_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_C_2")
  val ACC_C_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_C_3")
  val ACC_C_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_0")
  val ACC_H_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_1")
  val ACC_H_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_2")
  val ACC_H_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_H_3")
  val ACC_H_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_1_HI")
  val ARG_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_1_LO")
  val ARG_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_2_HI")
  val ARG_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_2_LO")
  val ARG_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BITS")
  val BITS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_NUM")
  val BIT_NUM: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_A_0")
  val BYTE_A_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_A_1")
  val BYTE_A_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_A_2")
  val BYTE_A_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_A_3")
  val BYTE_A_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_0")
  val BYTE_B_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_1")
  val BYTE_B_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_2")
  val BYTE_B_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_B_3")
  val BYTE_B_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_C_0")
  val BYTE_C_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_C_1")
  val BYTE_C_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_C_2")
  val BYTE_C_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_C_3")
  val BYTE_C_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_0")
  val BYTE_H_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_1")
  val BYTE_H_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_2")
  val BYTE_H_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_H_3")
  val BYTE_H_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COUNTER")
  val COUNTER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXPONENT_BIT")
  val EXPONENT_BIT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXPONENT_BIT_ACCUMULATOR")
  val EXPONENT_BIT_ACCUMULATOR: MutableList<String> = arrayListOf(),
  @get:JsonProperty("EXPONENT_BIT_SOURCE")
  val EXPONENT_BIT_SOURCE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INSTRUCTION")
  val INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MUL_STAMP")
  val MUL_STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OLI")
  val OLI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RESULT_VANISHES")
  val RESULT_VANISHES: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RES_HI")
  val RES_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RES_LO")
  val RES_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SQUARE_AND_MULTIPLY")
  val SQUARE_AND_MULTIPLY: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TINY_BASE")
  val TINY_BASE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TINY_EXPONENT")
  val TINY_EXPONENT: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Mxp(
  @get:JsonProperty("ACC_1")
  val ACC_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_2")
  val ACC_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_3")
  val ACC_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_4")
  val ACC_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_A")
  val ACC_A: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_Q")
  val ACC_Q: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_W")
  val ACC_W: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_1")
  val BYTE_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_2")
  val BYTE_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_3")
  val BYTE_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_4")
  val BYTE_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_A")
  val BYTE_A: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_Q")
  val BYTE_Q: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_QQ")
  val BYTE_QQ: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_R")
  val BYTE_R: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_W")
  val BYTE_W: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CN")
  val CN: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COMP")
  val COMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COUNTER")
  val COUNTER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("DELTA_MXPC")
  val DELTA_MXPC: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MAX_OFFSET")
  val MAX_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MAX_OFFSET_1")
  val MAX_OFFSET_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MAX_OFFSET_2")
  val MAX_OFFSET_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MEMORY_EXPANSION_EVENT")
  val MEMORY_EXPANSION_EVENT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MEMORY_EXPANSION_EXCEPTION")
  val MEMORY_EXPANSION_EXCEPTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXPC")
  val MXPC: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXPC_NEW")
  val MXPC_NEW: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_GBYTE")
  val MXP_GBYTE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_GWORD")
  val MXP_GWORD: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_INST")
  val MXP_INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_TYPE_1")
  val MXP_TYPE_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_TYPE_2")
  val MXP_TYPE_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_TYPE_3")
  val MXP_TYPE_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_TYPE_4")
  val MXP_TYPE_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MXP_TYPE_5")
  val MXP_TYPE_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NOOP")
  val NOOP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OFFSET_1_HI")
  val OFFSET_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OFFSET_1_LO")
  val OFFSET_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OFFSET_2_HI")
  val OFFSET_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OFFSET_2_LO")
  val OFFSET_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RIDICULOUSLY_OUT_OF_BOUND")
  val RIDICULOUSLY_OUT_OF_BOUND: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SIZE_1_HI")
  val SIZE_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SIZE_1_LO")
  val SIZE_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SIZE_2_HI")
  val SIZE_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SIZE_2_LO")
  val SIZE_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STAMP")
  val STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WORDS")
  val WORDS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WORDS_NEW")
  val WORDS_NEW: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class PhoneyRlp(
  @get:JsonProperty("INDEX")
  val INDEX: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LIMB")
  val LIMB: MutableList<String> = arrayListOf(),
  @get:JsonProperty("nBYTES")
  val N_BYTES: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TX_NUM")
  val TX_NUM: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Rlp(
  @get:JsonProperty("ADDR_HI")
  val ADDR_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ADDR_LO")
  val ADDR_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("addr_lo_1")
  val ADDR_LO_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("addr_lo_2")
  val ADDR_LO_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("addr_lo_ax")
  val ADDR_LO_AX: MutableList<String> = arrayListOf(),
  @get:JsonProperty("addr_lo_ndl")
  val ADDR_LO_NDL: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ct")
  val CT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("in-nonce")
  val IN_NONCE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NONCE")
  val NONCE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NONCE_ax")
  val NONCE_AX: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NONCE_bytes")
  val NONCE_BYTES: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NONCE_n")
  val NONCE_N: MutableList<String> = arrayListOf(),
  @get:JsonProperty("N_BYTES")
  val N_BYTES: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OUT")
  val OUT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("out2_shift")
  val OUT_2_SHIFT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("STAMP")
  val STAMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("tn")
  val TN: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Rom(
  @get:JsonProperty("ADDRESS_INDEX")
  var ADDRESS_INDEX: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CODEHASH_HI")
  var CODEHASH_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CODEHASH_LO")
  var CODEHASH_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CODESIZE")
  var CODESIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CODESIZE_REACHED")
  var CODESIZE_REACHED: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CODE_FRAGMENT_INDEX")
  var CODE_FRAGMENT_INDEX: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COUNTER")
  var COUNTER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CURRENT_CODEWORD")
  var CURRENT_CODEWORD: MutableList<String> = arrayListOf(),
  @get:JsonProperty("CYCLIC_BIT")
  var CYCLIC_BIT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_BYTECODE")
  var IS_BYTECODE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_INITCODE")
  var IS_INITCODE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_PUSH")
  var IS_PUSH: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_PUSH_DATA")
  var IS_PUSH_DATA: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OPCODE")
  var OPCODE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PADDED_BYTECODE_BYTE")
  var PADDED_BYTECODE_BYTE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PADDING_BIT")
  var PADDING_BIT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PC")
  var PC: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PUSH_FUNNEL_BIT")
  var PUSH_FUNNEL_BIT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PUSH_PARAMETER")
  var PUSH_PARAMETER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PUSH_PARAMETER_OFFSET")
  var PUSH_PARAMETER_OFFSET: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PUSH_VALUE_ACC_HI")
  var PUSH_VALUE_ACC_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PUSH_VALUE_ACC_LO")
  var PUSH_VALUE_ACC_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PUSH_VALUE_HI")
  var PUSH_VALUE_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PUSH_VALUE_LO")
  var PUSH_VALUE_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SC_ADDRESS_HI")
  var SC_ADDRESS_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SC_ADDRESS_LO")
  var SC_ADDRESS_LO: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Shf(
  @get:JsonProperty("ACC_1")
  val ACC_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_2")
  val ACC_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_3")
  val ACC_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_4")
  val ACC_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_5")
  val ACC_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_1_HI")
  val ARG_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_1_LO")
  val ARG_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_2_HI")
  val ARG_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARG_2_LO")
  val ARG_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BITS")
  val BITS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_1")
  val BIT_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_2")
  val BIT_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_3")
  val BIT_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_4")
  val BIT_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_B_3")
  val BIT_B_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_B_4")
  val BIT_B_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_B_5")
  val BIT_B_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_B_6")
  val BIT_B_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_B_7")
  val BIT_B_7: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_1")
  val BYTE_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_2")
  val BYTE_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_3")
  val BYTE_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_4")
  val BYTE_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_5")
  val BYTE_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COUNTER")
  val COUNTER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INST")
  val INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_DATA")
  val IS_DATA: MutableList<String> = arrayListOf(),
  @get:JsonProperty("KNOWN")
  val KNOWN: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LEFT_ALIGNED_SUFFIX_HIGH")
  val LEFT_ALIGNED_SUFFIX_HIGH: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LEFT_ALIGNED_SUFFIX_LOW")
  val LEFT_ALIGNED_SUFFIX_LOW: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LOW_3")
  val LOW_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MICRO_SHIFT_PARAMETER")
  val MICRO_SHIFT_PARAMETER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NEG")
  val NEG: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ONES")
  val ONES: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ONE_LINE_INSTRUCTION")
  val ONE_LINE_INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RES_HI")
  val RES_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RES_LO")
  val RES_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RIGHT_ALIGNED_PREFIX_HIGH")
  val RIGHT_ALIGNED_PREFIX_HIGH: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RIGHT_ALIGNED_PREFIX_LOW")
  val RIGHT_ALIGNED_PREFIX_LOW: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHB_3_HI")
  val SHB_3_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHB_3_LO")
  val SHB_3_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHB_4_HI")
  val SHB_4_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHB_4_LO")
  val SHB_4_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHB_5_HI")
  val SHB_5_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHB_5_LO")
  val SHB_5_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHB_6_HI")
  val SHB_6_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHB_6_LO")
  val SHB_6_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHB_7_HI")
  val SHB_7_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHB_7_LO")
  val SHB_7_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHIFT_DIRECTION")
  val SHIFT_DIRECTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("SHIFT_STAMP")
  val SHIFT_STAMP: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class ShfRt(
  @get:JsonProperty("BYTE")
  val BYTE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("IS_IN_RT")
  val IS_IN_RT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LAS")
  val LAS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("MSHP")
  val MSHP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ONES")
  val ONES: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RAP")
  val RAP: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class TxRlp(
  @get:JsonProperty("ABS_TX_NUM")
  val ABS_TX_NUM: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACCESS_TUPLE_BYTESIZE")
  val ACCESS_TUPLE_BYTESIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_1")
  val ACC_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_2")
  val ACC_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_BYTESIZE")
  val ACC_BYTESIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ADDR_HI")
  val ADDR_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ADDR_LO")
  val ADDR_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT")
  val BIT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_ACC")
  val BIT_ACC: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_1")
  val BYTE_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_2")
  val BYTE_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COMP")
  val COMP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COUNTER")
  val COUNTER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("DATAGASCOST")
  val DATAGASCOST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("DEPTH_1")
  val DEPTH_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("DEPTH_2")
  val DEPTH_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("DONE")
  val DONE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("end_phase")
  val END_PHASE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INDEX_DATA")
  val INDEX_DATA: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INDEX_LT")
  val INDEX_LT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INDEX_LX")
  val INDEX_LX: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INPUT_1")
  val INPUT_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INPUT_2")
  val INPUT_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("is_bytesize")
  val IS_BYTESIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("is_list")
  val IS_LIST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("is_padding")
  val IS_PADDING: MutableList<String> = arrayListOf(),
  @get:JsonProperty("is_prefix")
  val IS_PREFIX: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LIMB")
  val LIMB: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LIMB_CONSTRUCTED")
  val LIMB_CONSTRUCTED: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LT")
  val LT: MutableList<String> = arrayListOf(),
  @get:JsonProperty("LX")
  val LX: MutableList<String> = arrayListOf(),
  @get:JsonProperty("nb_Addr")
  val NB_ADDR: MutableList<String> = arrayListOf(),
  @get:JsonProperty("nb_Sto")
  val NB_STO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("nb_Sto_per_Addr")
  val NB_STO_PER_ADDR: MutableList<String> = arrayListOf(),
  @get:JsonProperty("number_step")
  val NUMBER_STEP: MutableList<String> = arrayListOf(),
  @get:JsonProperty("nBYTES")
  val N_BYTES: MutableList<String> = arrayListOf(),
  @get:JsonProperty("OLI")
  val OLI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_0")
  val PHASE_0: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_1")
  val PHASE_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_10")
  val PHASE_10: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_11")
  val PHASE_11: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_12")
  val PHASE_12: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_13")
  val PHASE_13: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_14")
  val PHASE_14: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_2")
  val PHASE_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_3")
  val PHASE_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_4")
  val PHASE_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_5")
  val PHASE_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_6")
  val PHASE_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_7")
  val PHASE_7: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_8")
  val PHASE_8: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_9")
  val PHASE_9: MutableList<String> = arrayListOf(),
  @get:JsonProperty("PHASE_BYTESIZE")
  val PHASE_BYTESIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("POWER")
  val POWER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RLP_LT_BYTESIZE")
  val RLP_LT_BYTESIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RLP_LX_BYTESIZE")
  val RLP_LX_BYTESIZE: MutableList<String> = arrayListOf(),
  @get:JsonProperty("TYPE")
  val TYPE: MutableList<String> = arrayListOf()
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Wcp(
  @get:JsonProperty("ACC_1")
  val ACC_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_2")
  val ACC_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_3")
  val ACC_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_4")
  val ACC_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_5")
  val ACC_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ACC_6")
  val ACC_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARGUMENT_1_HI")
  val ARGUMENT_1_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARGUMENT_1_LO")
  val ARGUMENT_1_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARGUMENT_2_HI")
  val ARGUMENT_2_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ARGUMENT_2_LO")
  val ARGUMENT_2_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BITS")
  val BITS: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_1")
  val BIT_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_2")
  val BIT_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_3")
  val BIT_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BIT_4")
  val BIT_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_1")
  val BYTE_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_2")
  val BYTE_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_3")
  val BYTE_3: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_4")
  val BYTE_4: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_5")
  val BYTE_5: MutableList<String> = arrayListOf(),
  @get:JsonProperty("BYTE_6")
  val BYTE_6: MutableList<String> = arrayListOf(),
  @get:JsonProperty("COUNTER")
  val COUNTER: MutableList<String> = arrayListOf(),
  @get:JsonProperty("INST")
  val INST: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NEG_1")
  val NEG_1: MutableList<String> = arrayListOf(),
  @get:JsonProperty("NEG_2")
  val NEG_2: MutableList<String> = arrayListOf(),
  @get:JsonProperty("ONE_LINE_INSTRUCTION")
  val ONE_LINE_INSTRUCTION: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RESULT_HI")
  val RESULT_HI: MutableList<String> = arrayListOf(),
  @get:JsonProperty("RESULT_LO")
  val RESULT_LO: MutableList<String> = arrayListOf(),
  @get:JsonProperty("WORD_COMPARISON_STAMP")
  val WORD_COMPARISON_STAMP: MutableList<String> = arrayListOf()
)

open class ConflatedTraceStorage(
  var hub: Hub = Hub(),
  var add: Add = Add(),
  var bin: Bin = Bin(),
  var binRt: BinRt = BinRt(),
  var ecData: EcData = EcData(),
  var ext: Ext = Ext(),
  var hashData: HashData = HashData(),
  var hashInfo: HashInfo = HashInfo(),
  var instructionDecoder: InstructionDecoder = InstructionDecoder(),
  var logData: LogData = LogData(),
  var logInfo: LogInfo = LogInfo(),
  var mmio: Mmio = Mmio(),
  var mmu: Mmu = Mmu(),
  var mmuId: MmuId = MmuId(),
  var mod: Mod = Mod(),
  var mul: Mul = Mul(),
  var mxp: Mxp = Mxp(),
  var phoneyRlp: PhoneyRlp = PhoneyRlp(),
  var rlp: Rlp = Rlp(),
  var rom: Rom = Rom(),
  var shf: Shf = Shf(),
  var shfRt: ShfRt = ShfRt(),
  var txRlp: TxRlp = TxRlp(),
  var wcp: Wcp = Wcp()
)
