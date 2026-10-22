package lineth.coordinator.clients.prover

import com.fasterxml.jackson.databind.DeserializationFeature
import com.fasterxml.jackson.databind.JsonNode
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.node.ObjectNode
import com.networknt.schema.JsonSchema
import com.networknt.schema.JsonSchemaFactory
import com.networknt.schema.SpecVersion
import lineth.coordinator.clients.prover.serialization.JsonSerialization
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Files
import java.nio.file.Path

/**
 * Checks the l2-execution DTOs against the rollup_spec wire contract: the golden fixtures under
 * `rollup_spec/src/rollup_spec/prover_io/testdata/` and the Draft 2020-12 schemas under `.../prover_io/schemas/`.
 *
 * Fixtures are read with `FAIL_ON_UNKNOWN_PROPERTIES` enabled (the production mapper disables it), so a field added
 * to the spec but missing from a DTO fails here instead of being silently dropped.
 */
class L2ExecutionProofSpecConformanceTest {
  private val proverIoDir: Path = Path.of(
    requireNotNull(System.getProperty("rollupSpecProverIoDir")) {
      "system property rollupSpecProverIoDir must point to rollup_spec/src/rollup_spec/prover_io"
    },
  )
  private val mapper: ObjectMapper = JsonSerialization.proofResponseMapperV1
  private val strictMapper: ObjectMapper = mapper.copy()
    .configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, true)

  private fun fixture(name: String): JsonNode = mapper.readTree(proverIoDir.resolve("testdata/$name").toFile())

  private fun schema(name: String): JsonSchema = Files.newInputStream(proverIoDir.resolve("schemas/$name")).use {
    JsonSchemaFactory.getInstance(SpecVersion.VersionFlag.V202012).getSchema(it)
  }

  /** Re-parses the serialized DTO so numbers compare by value (`IntNode` vs `LongNode`) against a parsed fixture. */
  private fun toJson(dto: Any): JsonNode = mapper.readTree(mapper.writeValueAsString(dto))

  private fun assertSchemaValid(schemaName: String, json: JsonNode) {
    assertThat(schema(schemaName).validate(json).map { it.toString() }).isEmpty()
  }

  @Test
  fun `l2-execution response fixture round-trips through the DTO and maps to the domain`() {
    val json = fixture("10-11-getZkL2ExecutionProofV1.response.json")

    val dto = strictMapper.treeToValue(json, L2ExecutionProofResponseDto::class.java)

    assertThat(toJson(dto)).isEqualTo(json)
    assertThat(dto.proverVersion).isEqualTo(json["proverVersion"].asText())
    val domain = L2ExecutionProofResponseDtoMapper(dto)
    assertThat(domain.endBlockNumber).isEqualTo(json["publicInputs"]["endBlockNumber"].asLong().toULong())
    assertThat(domain.publicInputs.blockCount).isEqualTo(json["publicInputs"]["blockCount"].asLong().toULong())
  }

  @Test
  fun `l2-execution request fixture round-trips through the DTO`() {
    val json = fixture("10-11-getZkL2ExecutionProofV1.request.json")

    val dto = strictMapper.treeToValue(json, L2ExecutionProofRequestDto::class.java)

    assertThat(toJson(dto)).isEqualTo(json)
  }

  @Test
  fun `l2-execution request built by the mapper is valid against the request schema`() {
    val dto = L2ExecutionProofRequestDtoMapper(
      guestProgramId = RiscvProverClientTestFixtures.L2_EXECUTION_PROGRAM_ID,
      provingSystem = RiscvProverClientTestFixtures.PROVING_SYSTEM_VERSION,
      l2MessageServiceAddress = RiscvProverClientTestFixtures.L2_MESSAGE_SERVICE_ADDRESS,
      forkName = RiscvProverClientTestFixtures.FORK_NAME,
    ).invoke(RiscvProverClientTestFixtures.l2ExecutionProofRequestV1()).get()

    val json = toJson(dto)
    assertSchemaValid("getZkL2ExecutionProofV1.request.schema.json", json)

    // Guard against a vacuous pass: the pre-spec envelope key must be rejected.
    val legacyEnvelope = (json.deepCopy<JsonNode>() as ObjectNode).apply {
      set<JsonNode>("programId", remove("guestProgramId"))
    }
    assertThat(schema("getZkL2ExecutionProofV1.request.schema.json").validate(legacyEnvelope)).isNotEmpty()
  }

  @Test
  fun `l2-execution proof embedded in a rollup request carries no proverVersion`() {
    val response = RiscvProverClientTestFixtures.l2ExecutionProofResponseDto(1000501L, 1000510L)
    assertThat(response.proverVersion).isNotNull()

    val dto = RollupProofRequestDtoMapper(
      guestProgramId = RiscvProverClientTestFixtures.ROLLUP_PROGRAM_ID,
      provingSystem = RiscvProverClientTestFixtures.PROVING_SYSTEM_VERSION,
      chainId = RiscvProverClientTestFixtures.CHAIN_ID,
      l2ExecutionProofProvider = { _ -> SafeFuture.completedFuture(response) },
    ).invoke(RiscvProverClientTestFixtures.rollupProofRequestV1()).get()

    val embedded = toJson(dto)["proofRequest"]["l2ExecutionProofs"]
    assertThat(embedded).hasSize(1)
    assertThat(embedded[0].has("proverVersion")).isFalse()
  }
}
