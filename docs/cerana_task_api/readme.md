Cerana Task Api
The Cerana Task api is an asynchronous Task System that breaks down tasks stacks by calling back to a common endpoint “Coordinator” instead of creating complex task stacks. This allows tasks to build upon tasks indefinitely as long as an api for a task is stable.

The Task Api is built on a basic communication library (acomm) that the entire system understands. Tasks are provided by Task Providers, and are registered by way of file system sockets. A Task Coordinator acts as a traffic manager for task providers, and doesn’t care what is providing the Task, it just uses the unix socket endpoint for a task.

Further documentation and understanding of the components and protocol can be obtained here:
- https://github.com/cerana/cerana/tree/master/acomm
- https://github.com/cerana/cerana/tree/master/provider
- https://github.com/cerana/cerana/tree/master/coordinator
