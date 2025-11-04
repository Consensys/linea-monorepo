package net.consensys.zkevm

import kotlin.time.Duration

/**
 *  * Interface representing a timer that can schedule and execute tasks at fixed intervals.
 *  * @param name The name of the timer.
 *  * @param task The task to be executed periodically. Can be blocking
 *  * @param initialDelay The initial delay before the first execution of the task.
 *  * @param period The period between successive executions of the task.
 *  * @param errorHandler A function to handle any errors that occur during task execution.
 *  */
interface LineaTimer {
  val name: String
  val task: Runnable
  val initialDelay: Duration
  val period: Duration
  val errorHandler: (Throwable) -> Unit
  fun start()
  fun stop()
}

interface TimerFactory {
  fun createTimer(
    name: String,
    task: Runnable,
    initialDelay: Duration,
    period: Duration,
    errorHandler: (Throwable) -> Unit,
  ): LineaTimer
}
