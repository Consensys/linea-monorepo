package parallel

type ThreadInit func(threadID int)
type Worker func(taskID, threadID int)
