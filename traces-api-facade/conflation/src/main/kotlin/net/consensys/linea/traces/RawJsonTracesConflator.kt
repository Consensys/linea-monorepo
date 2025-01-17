package net.consensys.linea.traces

import com.fasterxml.jackson.databind.MapperFeature
import com.fasterxml.jackson.databind.json.JsonMapper
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import io.vertx.core.json.JsonObject
import net.consensys.linea.ErrorType
import net.consensys.linea.TracesConflator
import net.consensys.linea.TracesError
import net.consensys.linea.VersionedResult
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.math.BigInteger
import kotlin.comparisons.compareBy
import kotlin.reflect.KProperty1
import kotlin.reflect.full.memberProperties

// Like .map(), but does so in place in the data structure
fun <T> MutableList<T>.mapInPlace(transform: (T) -> T) {
  for (i in this.indices) {
    this[i] = transform(this[i])
  }
}

// Like .map(), but does so in place in the data structure
fun <T> MutableList<T>.mapInPlaceWithIdx(transform: (Int, T) -> T) {
  for (i in this.indices) {
    this[i] = transform(i, this[i])
  }
}

// Given a class from modules.kt and one of it column, get & cast the actual data
inline fun <reified M : Any> getColumn(m: M, column: KProperty1<M, *>): MutableList<String> {
  @Suppress("UNCHECKED_CAST")
  return column.get(m)!! as MutableList<String>
}

// Simply append all the columns from the new block module
// to the existing column.
// This works for simple, non-stamped, modules.
inline fun <reified M : Any> appendTo(trace: M, target: M) {
  for (column in M::class.memberProperties) {
    getColumn(target, column).addAll(getColumn(trace, column))
  }
}

// Simply append all the columns from the new block module
// to the existing column.
// This works for simple, non-stamped, modules.
inline fun <reified M : Any> appendToMaybeSkip(trace: M, target: M, count: Int) {
  for (column in M::class.memberProperties) {
    if (getColumn(target, column).size > 0) {
      getColumn(target, column).addAll(getColumn(trace, column).drop(count))
    } else {
      getColumn(target, column).addAll(getColumn(trace, column))
    }
  }
}

// Append all the columns from the new block module to the
// existing column and re-stamp.
inline fun <reified M : Any> appendToRestamp(trace: M, target: M, stampName: String) {
  for (column in M::class.memberProperties) {
    if (column.name == stampName) {
      val currentStamp = getColumn(target, column).lastOrNull()?.toInt() ?: 0
      val reStamped = getColumn(trace, column).map { (it.toInt() + currentStamp).toString() }
      getColumn(target, column).addAll(reStamped)
    } else {
      getColumn(target, column).addAll(getColumn(trace, column))
    }
  }
}

data class HubState(
  var instructionStamp: Int = 0,
  var contextNumber: Int = 0,
  var txNum: Int = 0,
  var ramStamp: Int = 0,
  var stackStamp: Int = 0,
  var storageStamp: Int = 0,
  var hashNum: Int = 0,
  var logNum: Int = 0,
  var ecDataStamp: Int = 0
) {
  fun updateState(newHub: Hub) {
    newHub.INSTRUCTION_STAMP.mapInPlace { (it.toInt() + this.instructionStamp).toString() }
    this.instructionStamp = newHub.INSTRUCTION_STAMP.lastOrNull()?.toInt() ?: this.instructionStamp

    newHub.CONTEXT_NUMBER.mapInPlace { (it.toInt() + this.contextNumber).toString() }
    newHub.MAXIMUM_CONTEXT.mapInPlace { (it.toInt() + this.contextNumber).toString() }
    this.contextNumber = newHub.MAXIMUM_CONTEXT.lastOrNull()?.toInt() ?: this.contextNumber

    newHub.TX_NUM.mapInPlace { (it.toInt() + this.txNum).toString() }
    this.txNum = 1 + (newHub.TX_NUM.lastOrNull()?.toInt() ?: this.txNum)

    newHub.RAM_STAMP.mapInPlace { (it.toInt() + this.ramStamp).toString() }

    val restampStackStamp = { i: Int, x: String ->
      if (x.toInt() == 0 && newHub.STACK_STAMP[i].toInt() != 0) x else (x.toInt() + this.stackStamp).toString()
    }
    newHub.STACK_STAMP.mapInPlace { (it.toInt() + this.stackStamp).toString() }
    newHub.STACK_STAMP_NEW.mapInPlace { (it.toInt() + this.stackStamp).toString() }
    newHub.ITEM_STACK_STAMP_1.mapInPlaceWithIdx(restampStackStamp)
    newHub.ITEM_STACK_STAMP_2.mapInPlaceWithIdx(restampStackStamp)
    newHub.ITEM_STACK_STAMP_3.mapInPlaceWithIdx(restampStackStamp)
    newHub.ITEM_STACK_STAMP_4.mapInPlaceWithIdx(restampStackStamp)
    this.stackStamp = newHub.STACK_STAMP_NEW.lastOrNull()?.toInt() ?: this.stackStamp
  }
}

class ConflatedTrace : ConflatedTraceStorage() {
  private val log: Logger = LogManager.getLogger(this::class.java)

  var hubState = HubState()

  fun appendToHub(newHub: Hub) {
    this.hubState.updateState(newHub)
    appendTo(newHub, this.hub)
  }

  fun insertShfRt(shfRt: ShfRt) {
    this.shfRt = shfRt
  }

  fun insertBinRt(binRt: BinRt) {
    this.binRt = binRt
  }

  fun insertInstructionDecoder(id: InstructionDecoder) {
    this.instructionDecoder = id
  }

  fun insertMmuid(mmuId: MmuId) {
    this.mmuId = mmuId
  }

  fun fakeEcData() {
    this.ecData.cleanState()
  }

  fun fakeRom() {
    this.rom.ADDRESS_INDEX = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.CODEHASH_HI = mutableListOf(
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528",
      "262949717399590921288928019264691438528"
    )
    this.rom.CODEHASH_LO = mutableListOf(
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784",
      "304396909071904405792975023732328604784"
    )
    this.rom.CODESIZE = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.CODESIZE_REACHED = mutableListOf(
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1"
    )
    this.rom.CODE_FRAGMENT_INDEX = mutableListOf(
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1"
    )
    this.rom.COUNTER = mutableListOf(
      "0",
      "1",
      "2",
      "3",
      "4",
      "5",
      "6",
      "7",
      "8",
      "9",
      "10",
      "11",
      "12",
      "13",
      "14",
      "15",
      "0",
      "1",
      "2",
      "3",
      "4",
      "5",
      "6",
      "7",
      "8",
      "9",
      "10",
      "11",
      "12",
      "13",
      "14",
      "15"
    )
    this.rom.CURRENT_CODEWORD = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.CYCLIC_BIT = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1",
      "1"
    )
    this.rom.IS_BYTECODE = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.IS_INITCODE = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.IS_PUSH = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.IS_PUSH_DATA = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.OPCODE = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.PADDED_BYTECODE_BYTE = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.PADDING_BIT = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.PC = mutableListOf(
      "0",
      "1",
      "2",
      "3",
      "4",
      "5",
      "6",
      "7",
      "8",
      "9",
      "10",
      "11",
      "12",
      "13",
      "14",
      "15",
      "16",
      "17",
      "18",
      "19",
      "20",
      "21",
      "22",
      "23",
      "24",
      "25",
      "26",
      "27",
      "28",
      "29",
      "30",
      "31"
    )
    this.rom.PUSH_FUNNEL_BIT = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.PUSH_PARAMETER = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.PUSH_PARAMETER_OFFSET = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.PUSH_VALUE_ACC_LO = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.PUSH_VALUE_ACC_HI = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.PUSH_VALUE_LO = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.PUSH_VALUE_HI = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.SC_ADDRESS_LO = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
    this.rom.SC_ADDRESS_HI = mutableListOf(
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0",
      "0"
    )
  }

  fun reAssembleRom() {
    fun hiLoToAddr(hi: String, lo: String): BigInteger {
      val addrShift = 256.toBigInteger().pow(16)
      return hi.toBigInteger().multiply(addrShift).add(lo.toBigInteger())
    }

    val idxs = (0 until this.rom.PC.size).toList()
    val sortedIdxs = idxs.sortedWith(
      compareBy(
        { hiLoToAddr(this.rom.SC_ADDRESS_HI[it], this.rom.SC_ADDRESS_LO[it]) },
        // we want the initcode *FIRST*
        { -this.rom.IS_INITCODE[it].toInt() },
        { this.rom.PC[it].toInt() }
      )
    )

    // Awfully suboptimal, but it just works
    val copiedRom = this.rom.copy()
    this.rom = Rom()
    for (i in 0 until copiedRom.PC.size) {
      // Avoid writing twice the same contract
      var duplicate = true
      if (i > 0) {
        for (column in Rom::class.memberProperties) {
          if (column != Rom::ADDRESS_INDEX && column != Rom::CODE_FRAGMENT_INDEX) {
            duplicate = duplicate && getColumn(copiedRom, column)[sortedIdxs[i]].toBigInteger()
              .equals(getColumn(this.rom, column).last().toBigInteger())
          }
        }
      } else {
        duplicate = false
      }

      if (!duplicate) {
        for (column in Rom::class.memberProperties) {
          getColumn(this.rom, column).add(getColumn(copiedRom, column)[sortedIdxs[i]])
        }
      }
    }

    var currentAddress = hiLoToAddr(
      this.rom.SC_ADDRESS_HI[0],
      this.rom.SC_ADDRESS_LO[0]
    )
    var currentIsDeployment = this.rom.IS_INITCODE[0].toUInt()
    var codeFragmentIndex = 1
    var addressIndex = 0
    for (i in 0 until this.rom.PC.size) {
      val newIsDeployment = this.rom.IS_INITCODE[i].toUInt()
      var newAddress = hiLoToAddr(
        this.rom.SC_ADDRESS_HI[i],
        this.rom.SC_ADDRESS_LO[i]
      )

      if (newAddress != currentAddress) {
        currentAddress = newAddress
        currentIsDeployment = newIsDeployment

        addressIndex += 1
        codeFragmentIndex += 1
      } else if (newIsDeployment != currentIsDeployment) {
        currentIsDeployment = newIsDeployment
        codeFragmentIndex += 1
      }

      this.rom.ADDRESS_INDEX[i] = addressIndex.toString()
      this.rom.CODE_FRAGMENT_INDEX[i] = codeFragmentIndex.toString()
    }
  }

  fun add(module: Any) {
    when (module) {
      is Add -> appendToRestamp(module, this.add, "STAMP")
      is Bin -> appendToRestamp(module, this.bin, "BINARY_STAMP")
      is BinRt -> this.insertBinRt(module)
      is Ext -> appendToRestamp(module, this.ext, "STAMP")
      is EcData -> {
        module.STAMP.mapInPlace { it -> (it.toInt() + this.hubState.ecDataStamp).toString() }
        this.hubState.ecDataStamp = module.STAMP.lastOrNull()?.toInt() ?: this.hubState.ecDataStamp
        appendTo(module, this.ecData)
      }

      is HashData -> {
        appendToRestamp(module, this.hashData, "NUM")
        this.hubState.hashNum = this.hashData.NUM.lastOrNull()?.toInt() ?: this.hubState.hashNum
      }

      is HashInfo -> appendToRestamp(module, this.hashInfo, "HASH_NUM")
      is Hub -> appendToHub(module)
      is InstructionDecoder -> this.insertInstructionDecoder(module)
      is LogData -> {
        appendToRestamp(module, this.logData, "NUM")
        this.hubState.logNum = this.hashData.NUM.lastOrNull()?.toInt() ?: this.hubState.logNum
      }

      is LogInfo -> appendToRestamp(module, this.logInfo, "NUM")
      is Mmio -> {
        module.TX_NUM.mapInPlace { it -> (it.toInt() + this.hubState.txNum).toString() }
        appendTo(module, this.mmio)
      }

      is Mmu -> {
        appendToRestamp(module, this.mmu, "RAM_STAMP")
        this.hubState.ramStamp = 1 + (this.mmu.RAM_STAMP.lastOrNull()?.toInt() ?: this.hubState.ramStamp)
      }

      is MmuId -> this.insertMmuid(module)
      is Mod -> appendToRestamp(module, this.mod, "STAMP")
      is Mul -> appendToRestamp(module, this.mul, "MUL_STAMP")
      is Mxp -> {
        module.CN.mapInPlace { it -> (it.toInt() + this.hubState.contextNumber).toString() }
        appendToRestamp(module, this.mxp, "STAMP")
      }

      is PhoneyRlp -> {
        module.TX_NUM.mapInPlace { it -> (it.toInt() + this.hubState.txNum).toString() }
        appendTo(module, this.phoneyRlp)
      }

      is Rlp -> appendToRestamp(module, this.rlp, "STAMP")
      is Rom -> {} // appendToMaybeSkip(module, this.rom, 32)
      is Shf -> appendToRestamp(module, this.shf, "SHIFT_STAMP")
      is ShfRt -> this.insertShfRt(module)
      is TxRlp -> {
        appendToRestamp(module, this.txRlp, "ABS_TX_NUM")
      }

      is Wcp -> appendToRestamp(module, this.wcp, "WORD_COMPARISON_STAMP")
      else -> log.warn("unknown module {}", module::class.simpleName)
    }
  }
}

class RawJsonTracesConflator(private val tracesEngineVersion: String) : TracesConflator {
  private val objectMapper: JsonMapper = JsonMapper.builder().disable(MapperFeature.USE_GETTERS_AS_SETTERS).build()

  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun conflateTraces(
    traces: List<JsonObject>
  ): Result<VersionedResult<JsonObject>, TracesError> {
    return if (traces.isEmpty()) {
      Err(TracesError(ErrorType.TRACES_UNAVAILABLE, "Empty list of traces for conflation"))
    } else {
      val ax = ConflatedTrace()

      traces.forEach { trace ->
        // XXX: it is *CAPITAL* for the hub to be the first in the list
        // so that it can update its stamps
        MODULES.forEach { (jsonPath, klass) ->
          log.trace("Parsing trace: {}", jsonPath)
          trace.getTrace(jsonPath)?.let {
            if (!it.isEmpty) {
              ax.add(objectMapper.convertValue(it, klass))
            }
          } ?: run {
            log.warn("Could not parse object with path: '{}'", jsonPath.joinToString("."))
          }
        }
      }

      ax.fakeRom()
      // Geth EC_DATA arithmetization is erroneous and may overflow the field; disable it.
      ax.fakeEcData()
      val result = JsonObject()
      result.put("add", JsonObject.of("Trace", ax.add))
      result.put("bin", JsonObject.of("Trace", ax.bin))
      result.put("binRT", JsonObject.of("Trace", ax.binRt))
      result.put("ext", JsonObject.of("Trace", ax.ext))
      result.put("ec_data", JsonObject.of("Trace", ax.ecData))
      result.put("hub", JsonObject.of("Trace", ax.hub))
      result.put("instruction-decoder", JsonObject.of("Trace", ax.instructionDecoder))
      result.put("mmio", JsonObject.of("Trace", ax.mmio))
      result.put("mmu", JsonObject.of("Trace", ax.mmu))
      result.put("mmuID", JsonObject.of("Trace", ax.mmuId))
      result.put("mod", JsonObject.of("Trace", ax.mod))
      result.put("mul", JsonObject.of("Trace", ax.mul))
      result.put("mxp", JsonObject.of("Trace", ax.mxp))
      result.put(
        "pub",
        JsonObject.of(
          "hash_data",
          JsonObject.of("Trace", ax.hashData),
          "hash_info",
          JsonObject.of("Trace", ax.hashInfo),
          "log_data",
          JsonObject.of("Trace", ax.logData),
          "log_info",
          JsonObject.of("Trace", ax.logInfo)
        )
      )
      result.put("phoneyRLP", JsonObject.of("Trace", ax.phoneyRlp))
      result.put("rlp", JsonObject.of("Trace", ax.rlp))
      result.put("rom", JsonObject.of("Trace", ax.rom))
      result.put("shf", JsonObject.of("Trace", ax.shf))
      result.put("shfRT", JsonObject.of("Trace", ax.shfRt))
      result.put("txRlp", JsonObject.of("Trace", ax.txRlp))
      result.put("wcp", JsonObject.of("Trace", ax.wcp))

      Ok(VersionedResult(tracesEngineVersion, result))
    }
  }
}
