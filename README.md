# DESIGN

## MOTIVATION
The purpose of the scheduling framework is to schedule user-defined tasks,
support periodic scheduling and error retry,and support task persistence.
## ARCHITECTURE

![img.png](img.png)

### JOB

db table design

| Column |  Type | Description |
| --- | --- | --- |
|   id  |  bigint  | primary key  |
|   name  |  bigint  | job name  |
|   retry_times  |  bigint  | number of job retries when execution fails  |
|   runonce  |  tinyint  | whether the job is run only once  |
|   interval  |  varchar  | job run interval  |
|   created_time  |  datetime  | job created time |
|   updated_time  |  datetime  | job updated time |
|   lastFinished_time  |  datetime  | job last finished time  |
|   is_delete  |  tinyint  | logic delete flag  |

### WorkQueue

WorkQueue support put item into a queue delayed that achieve run jobs periodically.

### Processor

Processor get a job from workQueue that is core in the framework,can run multiple Worker.

### Worker

Worker run logic every period, run job's task func and update metadata.

## TEST

> go run main.go

### CASE

1. task1: 10-second interval task
2. task2: runonce task
3. task3: 3-retries task

```go
// task1 interval job
func task1() error {
log.Info("task1 running.....")

return nil
}

// task2 runOnce job
func task2() error {
log.Info("task2 running.....")

return nil
}

// task3 error job
func task3() error {
log.Info("task3 running.....")

return fmt.Errorf("mock job faild")
}

...

job1 := job.NewJob("task1").AddFunc(task1).
SetRunOnce(false).SetInterval(10 * time.Second)

job2 := job.NewJob("task2").AddFunc(task2).
SetRunOnce(true)

job3 := job.NewJob("task3").AddFunc(task3).
SetRunOnce(true).SetRetry(3)
```

### CASE RESULT

```shell
{"level":"info","msg":"task1 running.....","time":"2022-06-25T17:40:41+08:00"}
{"level":"info","msg":"task2 running.....","time":"2022-06-25T17:40:41+08:00"}
{"level":"info","msg":"task3 running.....","time":"2022-06-25T17:40:41+08:00"}
{"level":"info","msg":"task3 running.....","time":"2022-06-25T17:40:41+08:00"}
{"level":"info","msg":"task3 running.....","time":"2022-06-25T17:40:41+08:00"}
{"level":"info","msg":"task3 running.....","time":"2022-06-25T17:40:42+08:00"}
{"level":"info","msg":"task1 running.....","time":"2022-06-25T17:40:51+08:00"}
{"level":"info","msg":"task1 running.....","time":"2022-06-25T17:41:01+08:00"}
{"level":"info","msg":"task1 running.....","time":"2022-06-25T17:41:11+08:00"}
{"level":"info","msg":"task1 running.....","time":"2022-06-25T17:41:21+08:00"}

```