package linea.timer

import kotlin.time.Duration

enum class TimerSchedule {
  FIXED_DELAY,
  FIXED_RATE,
}

/**
 *  * Interface representing a timer that can schedule and execute tasks at fixed intervals.
 *  * @param name The name of the timer.
 *  * @param task The task to be executed periodically. Can be blocking
 *  * @param initialDelay The initial delay before the first execution of the task.
 *  * @param period The period between successive executions of the task in case of Fixed Delay
 *  * and average frequency in case of fixed rate timers.
 *  * @param errorHandler A function to handle any errors that occur during task execution.
 *  */
interface Timer {
  val name: String
  val task: Runnable
  val initialDelay: Duration
  val period: Duration
  val errorHandler: (Throwable) -> Unit
  val timerSchedule: TimerSchedule
  fun start()
  fun stop()
}

interface TimerFactory {
  fun createTimer(
    name: String,
    initialDelay: Duration,
    period: Duration,
    timerSchedule: TimerSchedule,
    errorHandler: (Throwable) -> Unit,
    task: Runnable,
  ): Timer
}
