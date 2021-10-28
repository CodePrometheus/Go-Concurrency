# Lab1

[TOC]



初始代码默认已经提供了 单进程串行 的 MapReduce 参考实现，可通过如下命令先跑一个简单的 Demo

~~~go
# 构建 MapReduce APP 的动态链接库
go build -race -buildmode=plugin ../mrapps/wc.go

# 运行 MapReduce
rm mr-out*
go run -race mrsequential.go wc.so pg*.txt

# 查看结果
more mr-out-0
~~~

> 这里使用了 Golang 的 Plugin 来构建 MR APP，使得 MR 框架的代码可以和 MR APP 的代码分开编译，而后 MR 框架再通过动态链接的方式载入指定的 MR APP 运行。






## 任务分析

需要实现的是一个 单机多进程并行 的 MapReduce

- 整个 MR 框架由一个 Coordinator 进程及若干个 Worker 进程构成
- Coordinator 进程与 Worker 进程间通过**本地 Socket** 进行 Golang RPC 通信
- 由 Coordinator 协调整个 MapReduce 计算的推进，并**分配 Task 到 Worker** 上运行
- 在启动 Coordinator 进程时指定 输入文件名 及 Reduce Task 数量
- 在启动 Worker 进程时指定所用的 MR APP 动态链接库文件
- Coordinator 需要留意 Worker 可能无法在合理时间内完成收到的任务（Worker 卡死或宕机），在遇到此类问题时**需要重新派发任务**
- Coordinator 进程的入口文件为 main/mrcoordinator.go
- Worker 进程的入口文件为 main/mrworker.go
- 需要补充实现 mr/coordinator.go、mr/worker.go、mr/rpc.go 这三个文件





## Coordinator 功能

- 在启动时根据 指定的输入文件数 及 Reduce Task 数，**生成 Map Task 及 Reduce Task**
- 响应 Worker 的 Task 申请 RPC 请求，同步阻塞，直到**分配可用的 Task** 给该 Worker 处理
- 追踪 Task 的完成情况，在 Task 超出10s仍未完成时，将该 Task 重新分配给其他 Worker 重试，在所有 Map Task 完成后进入 Reduce 阶段，**开始派发 Reduce Task**；在所有 Reduce Task 完成后标记作业已完成并退出
- 考虑 Task 上一次分配的 Worker 可能仍在运行，重新分配后会出现两个 Worker 同时运行同一个 Task 的情况。要**确保只有一个 Worker 能够完成结果数据的最终写出**，以免出现冲突，导致下游观察到重复或缺失的结果数据



实现 Coordinator   最后一点，可参考 Google MapReduce 的做法，Worker 在写出数据时可以先写出到临时文件，最终确认没有问题后再将其重命名为正式结果文件，**区分开了 Write 和 Commit 的过程**。Commit 的过程可以是 Coordinator 来执行，也可以是 Worker 来执行：

- Coordinator Commit：Worker 向 Coordinator 汇报 Task 完成，**Coordinator 确认该 Task 是否仍属于该 Worker，是则进行结果文件 Commit，否则直接忽略**
- Worker Commit：Worker 向 Coordinator 汇报 Task 完成，Coordinator 确认该 Task 是否仍属于该 Worker 并响应 Worker，是则 Worker 进行结果文件 Commit，**再向 Coordinator 汇报 Commit 完成**

这里两种方案都是可行的，各有利弊。实现中选择了 Coordinator Commit，因为它可以少一次 RPC 调用，在编码实现上会更简单，但缺点是所有 Task 的最终 Commit 都由 Coordinator 完成，在极端场景下会让 Coordinator 变成整个 MR 过程的性能瓶颈





## Worker 功能

- Worker 首先应判断当前任务是 Map 还是 Reduce 类型
- 如果是 Map 类型，需要知道该任务要处理的文件名，然后用自己的 mapf 函数读取该文件中的内容并进行计算，返回一个 KeyValue 的切片，将得到的结果 存在文件中等待 reduce 处理，并且一个 Map Task 要将结果存成 nReduce 份（nReduce是reduce任务的数量），并且给文件命名中包含对应要处理某个中间结果的reducerID以让之后的reduce任务知道自己要读取哪些中间结果
- 如果是 Reduce 类型，就根据自己的 Reduce TaskID，读取对应的中间结果文件，然后调用 Reduce 函数计算得到最终的结果
- 在空闲的时候通过 RPC 向 Coordinator  申请 Task 并运行，再不断重复这个过程即可，Worker是可以多个并行执行的
- Worker 在上一个 Task 完成后 向 Coordinator  汇报
- 合并上述的两个功能，由 Worker 向 Coordinator  申请一个新的 Task，同时汇报上一个运行完成的 Task（如有）
- 考虑 Worker 的 Failover，即 Worker 获取到 Task 后可能出现宕机和卡死等情况。这两种情况在 Coordinator 的视角中都是相同的，就是该 Worker 长时间不与 Coordinator 通信了。为了简化任务，Lab 说明中明确指定了，设定该超时阈值为 10s 即可





## 扩展与展望

假设虽然超过10s后，master以为该worker crash掉了，但是可能只是该worker这次处理慢了而已，可能15s后就向master报告自己已经处理完成。这种情况下，master应该要拒绝将临时结果的文件名重命名；因此我在Task任务结构体中还增加了一个FailedWorkers切片变量，用于存放master认为无法完成该任务的workerID，对于不再该切片中的worker的完成报告才给予接受和重命名。
