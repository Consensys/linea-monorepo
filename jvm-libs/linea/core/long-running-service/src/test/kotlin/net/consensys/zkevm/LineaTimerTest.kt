package net.consensys.zkevm

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.MethodSource
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicInteger
import kotlin.concurrent.atomics.AtomicBoolean
import kotlin.concurrent.atomics.AtomicReference
import kotlin.concurrent.atomics.ExperimentalAtomicApi
import kotlin.random.Random
import kotlin.reflect.KClass
import kotlin.time.Clock
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.ExperimentalTime
import kotlin.time.toJavaDuration

@OptIn(ExperimentalAtomicApi::class)
class LineaTimerTest {
  companion object {
    @JvmStatic
    fun timerTypes() = listOf(
      JvmTimer::class,
      VertxTimer::class,
    )

    fun timerReference(timer: LineaTimer): Any? {
      return when (timer) {
        is VertxTimer -> timer.timerReference()
        is JvmTimer -> timer.timerReference()
        else -> throw IllegalArgumentException("Unknown timer type")
      }
    }
  }

  private val vertx = Vertx.vertx()

  fun createTimer(
    timerType: KClass<out LineaTimer>,
    task: Runnable,
    initialDelay: Duration = 50.milliseconds,
    period: Duration = 50.milliseconds,
    errorHandler: (Throwable) -> Unit,
  ): LineaTimer {
    return when (timerType) {
      JvmTimer::class -> JvmTimer(
        name = "test-jvm-timer-${Random.nextInt()}",
        task = task,
        initialDelay = initialDelay,
        period = period,
        errorHandler = errorHandler,
      )
      VertxTimer::class -> VertxTimer(
        vertx = vertx,
        name = "test-vertx-timer-${Random.nextInt()}",
        task = task,
        initialDelay = initialDelay,
        period = period,
        errorHandler = errorHandler,
      )
      else -> throw IllegalArgumentException("Unknown timer type: $timerType")
    }
  }

  @ParameterizedTest
  @MethodSource("timerTypes")
  @Timeout(2, timeUnit = TimeUnit.SECONDS)
  fun `task exception is handled by error handler`(typerType: KClass<out LineaTimer>) {
    val errorHandled = AtomicBoolean(false)
    val error = AtomicReference<String>("")
    val timer = createTimer(
      timerType = typerType,
      task = Runnable { throw RuntimeException("Test error") },
      errorHandler = { t ->
        errorHandled.store(true)
        error.store(t.message!!)
      },
    )
    timer.start()
    await()
      .pollInterval(30.milliseconds.toJavaDuration())
      .untilAsserted {
        assertThat(errorHandled.load()).isTrue()
        assertThat(error.load()).isEqualTo("Test error")
      }
    timer.stop()
  }

  @ParameterizedTest
  @MethodSource("timerTypes")
  @Timeout(2, timeUnit = TimeUnit.SECONDS)
  fun `timer continues to tick after exception`(typerType: KClass<out LineaTimer>) {
    val latch = CountDownLatch(5)
    val timer = createTimer(
      timerType = typerType,
      task = Runnable {
        latch.countDown()
        throw RuntimeException("Test error")
      },
      errorHandler = { },
    )
    timer.start()
    assertThat(latch.await(5, TimeUnit.SECONDS)).isTrue()
    timer.stop()
  }

  @OptIn(ExperimentalTime::class)
  @ParameterizedTest
  @MethodSource("timerTypes")
  @Timeout(2, timeUnit = TimeUnit.SECONDS)
  fun `ticks shouldn't run concurrently if execution is blocked for more than polling interval`(
    typerType: KClass<out LineaTimer>,
  ) {
    val pollingInterval = 50.milliseconds
    val numberOfInvocations = AtomicInteger(0)
    val prevActionIsRunning = AtomicBoolean(false)
    val actionWasCalledInParallel = AtomicBoolean(false)
    val timer = createTimer(
      timerType = typerType,
      task = Runnable {
        numberOfInvocations.incrementAndGet()
        if (prevActionIsRunning.load()) {
          actionWasCalledInParallel.store(true)
        }
        prevActionIsRunning.store(true)
        val executionStartTime = Clock.System.now()
        while (Clock.System.now() - executionStartTime < pollingInterval.times(3)) {
          Thread.sleep(10)
        }
        prevActionIsRunning.store(false)
      },
      errorHandler = { },
    )
    timer.start()
    await()
      .pollInterval(30.milliseconds.toJavaDuration())
      .untilAsserted {
        println("$typerType: " + numberOfInvocations.get() + ", " + actionWasCalledInParallel.load())
        assertThat(numberOfInvocations.get()).isGreaterThanOrEqualTo(5)
        assertThat(actionWasCalledInParallel.load()).isFalse()
      }
    timer.stop()
  }

  @ParameterizedTest
  @MethodSource("timerTypes")
  fun `timer start is idempotent`(typerType: KClass<out LineaTimer>) {
    val timerReferences = mutableListOf<Any?>()
    val timer = createTimer(
      timerType = typerType,
      task = Runnable { },
      errorHandler = { },
    )
    repeat(5) {
      timer.start()
      timerReferences.add(timerReference(timer))
    }
    val firstReference = timerReferences.first()
    timerReferences.forEach {
      assertThat(it).isEqualTo(firstReference)
    }
  }
}
