## Overview
The task subsystem is designed so a single node can easily accomplish anything by breaking down goals into smaller subsets of tasks (which may in turn be broken down into subsets of tasks).
The task subsystem's ultimate goal is to provide an easy way to break down units of functionality in a way that many pieces of code can provide one or more tasks and use them in other tasks by way of task coordination.

## Tasks
Tasks are the basic building block of this entire system. Tasks are designed to do a thing, regardless if that thing is just give information, or change the state of the current node. Tasks can use other tasks. Example:
* Request for getnodestats
  * Request for getcpu
  * Request for getfreemem
  * Request for getallmem
  * etc....
Tasks are designed to provide success or failure after what they attempt to do, and all tasks should have procedures to follow in the event of both.

## The Code
There are two types of software involved in the system: The Task Coordinator (only one), and Task Providers (one or more). All pieces of code run conceptually as containers, but with a specially bind mounted socket directory where they can interact locally through sockets **(/task-sockets/$provider_$priority_$taskname.socket)**.

### Task Coodinator
The task coordinator's simple job is to allow anything to request a task, and get data from the outcome. The task coordinator listens two ways:
* TCP port 7050
* /task-sockets/task-coordinator_100_listener.socket

When a task request comes in, the task coordinator looks for the socket best suited to handle that task, which is the socket where $priority is highest and $taskname matches the task. Example: 
* If a task **getcpu** is requested, and sockets **/task-sockets/generic-provider_50_getcpu.socket** and **/task-sockets/better-provider_60_getcpu.socket** exist, the task-coordinator will forward the task request only to **/task-sockets/better-provider_60_getcpu.socket**

The task coordinator is not only meant for intra-task communication, but for initial requests to also come from an outside source over the TCP listener. When this happens, it simply maps a path back to the external listener for the tasks to use (to be described in detail in the communication section)

### Task Providers
Task providers provide one or more tasks for the system to use. Each task has it's own unique socket, so, for example, a provider with 5 tasks should provide 5 sockets. Tasks can make requests to the task coordinator as well, so they can use the results of other tasks.

Task provider software will be run in an environment with no network access. The only means of interaction they have are sockets. When a task request comes in, it comes with a statement of where the response should go, and that, to a task provider, should always be another socket. If the originator was a tcp request, the task-coordinator provides a socket for it to proxy through.

## Communication
Communication between sockets uses json payloads. Tasks are always initated with task requests (which are replied to immediately with either an ok or a reason the task cannot be started), and a task response is given later when the task is completed.

### Task request
```javascript
{
  id:    "string",
    // unique identifier that makes sense to the requestor
  task:  "string",
    // name of task
  responseURL: "string",
    // location where response should be sent (http:// for outside and unix:// for inside)
  args: "Freeform, probably object of named keys"
    // the arguments that matter to a task provider
}
```
### Task request adknowledgement
```javascript
{
  success: "boolean"
    // true if task startable, false if not
  message: "string"
    // optional message for reason of success or failure
}
```
### Task response
```javascript
{
  id: "string",
    // the id that came in with the request
  result: "freeform",
    // null if unsuccessful, data that the requestor cares about otherwise
  error: "freeform, likely object",
    // null if successful, populated otherwise (TODO: define this more strictly)
  streamURL: "string",
    // optional url if there is raw data for the requestor to stream (ex: zfs snapshot)
}
```
