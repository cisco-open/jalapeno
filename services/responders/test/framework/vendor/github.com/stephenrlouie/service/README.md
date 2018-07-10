# Service

[![Build Status](https://travis-ci.org/stephenrlouie/service.svg?branch=master)](https://travis-ci.org/stephenrlouie/service)
[![GoDoc](https://godoc.org/github.com/stephenrlouie/service?status.png)](https://godoc.org/github.com/stephenrlouie/service)


A utility to standardize and simplify managing go routines in a go program. It's intent is to handle channels and closing all ServiceGroups specified if one of the contained services fails.

Services could be long running or short running routines.

## Interface

Responsibilities of a Service

1. Log all errors
2. Run until the completion of the service
3. Only return fatal errors to the ServiceGroup
4. Return `nil` on successful execution or graceful exit

`func Start() error` : The entry point of the service, should not return until the service is done.

`func Stop()`: The graceful stop function that should take a thread safe operation on the implementation of `Service` to cause the `Start` function to exit.

## ServiceGroup

A list of services to start, watch for errors and properly exit all services in the same ServiceGroup. There should be some dependence between services in a ServiceGroup since a fatal error in one of the services causes the entire ServiceGroup to shut down.

- `New` : creates a new ServiceGroup
- `Add` : adds a service to the ServiceGroup
- `Wait`: ensures the main thread will block until all services in the ServiceGroup are done running
- `Kill`: force close everything in a ServiceGroup
- `Start`: starts all the services in the ServiceGroup
- `HandleSigint`: adds a handler to stop all services on SIGINT


## Other Notes

Each child routine is expected to return one error at most. Internally each routine's error channel is merged into an error channel that has a capacity equal to the number of child routines. If the child routines pass more errors than the number of child routines the channel will fill up and crash the program.


## Examples

Please see [examples](https://github.com/stephenrlouie/service/tree/master/examples). I provide an example of short running services and long running services and a failure case.

We have two shorter running services `hello` and `sleep-2` that pass. `sleep-4` will fail, causing a ServiceGroup shutdown, `sleep-6` and `sleep-8` are still running and will be shut down, `sleep-8` will fail upon the forced shutdown and `sleep-6` will gracefully close. You can see how errors are captured in `main.go`

```
service $go run examples/main.go
hello says: 'Hello world'
sleep-4!
sleep-6!
sleep-2!
sleep-8!
sleep-4!
sleep-6!
sleep-8!
sleep-2!
sleep-6!
sleep-2 is closed
sleep-8!
sleep-4!
sleep-4!
sleep-6!
sleep-8!
sleep-4 is closed
sleep-6!
sleep-8!
hello stop
sleep-2 stop
sleep-4 stop
sleep-6 stop
sleep-8 stop
sleep-6 is closed
sleep-8 is closed
*** Service Group Errors ***
        0: sleep-4 fail
        1: sleep-8 fail
```
