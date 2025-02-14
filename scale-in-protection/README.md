# Protect EC2 instances from being scaled down while busy

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/scale-in-protection.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/scale-in-protection)

Monitor workers' statuses to enable or disable instance scale-in protection accordingly.

## Inspiration

The need for this monitor arises
from https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-using-sqs-queue.html#scale-sqs-queue-scale-in-protection:

```java
while (true)
{
    SetInstanceProtection(False);
    Work = GetNextWorkUnit();
    SetInstanceProtection(True);
    ProcessWorkUnit(Work);
    SetInstanceProtection(False);
}
```

Essentially, if you have any number of workers who can be either ACTIVE or IDLE, you generally want to enable scale-in
protection when any of your worker is actively doing some work, while once all the workers have become idle, you would
want to disable scale-in protection to let the Auto Scaling group reclaim your instance naturally.

**Note**: there is a possibility that your instance is terminated in-between the `GetWorkUnit()` and the
`ProcessWorkUnit(Work)` calls. Generally if your visibility timeout is low enough, this is not an issue as a different
worker would be able to pick up the message again.

## Setup

```bash
go get github.com/nguyengg/go-scale-in-protection
```

## Example usage

There is only one very simple `ScaleInProtector` struct. You can declare a dependency on the package and use it; and I
would recommend copying, pasting, and customising it to your needs as well.

```go
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	sip "github.com/nguyengg/go-aws-commons/scale-in-protection"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := &sip.ScaleInProtector{
		// both InstanceId and AutoScalingGroupName are optional.
		// the monitor knows how to use the default aws.Config to figure out its own instance Id (via IMDSv2) and the
		// Auto Scaling group name containing that instance id.
		InstanceId:           "i-1234",
		AutoScalingGroupName: "my-asg",
		// when the last worker becomes idle, wait for another 15" before marking the instance as safe to terminate.
		IdleAtLeast: 15 * time.Minute,
	}

	// you can start the workers first or start the monitor's main loop first, either works.
	// the monitor doesn't need to know beforehand how many workers there are but you should signal active at least once
	// because the monitor starts out assuming all workers are idle.
	workerCount := 5
	var wg sync.WaitGroup
	for i := range workerCount {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			for {
				// get some work.

				// mark active.
				s.SignalActive(id)

				// do the work.

				// mark idle.
				s.SignalIdle(id)
			}
		}(strconv.Itoa(i))
	}

	// always starts the monitor's main loop in a goroutine because it doesn't return until context is cancelled or it
	// encounters an error.
	wg.Add(1)
	go func() {
		if err := s.StartMonitoring(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			log.Fatal(err)
		}
	}()

	wg.Wait()
}

```
