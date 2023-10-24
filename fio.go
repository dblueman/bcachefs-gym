package main

import (
   "fmt"
   "sync/atomic"
)

var (
   workloadActive atomic.Bool
)

func workload() error {
   args := []string{"fio",
      "--group_reporting", "--ioengine=io_uring", "--directory=" + mountPath,
      "--size=16m", "--time_based", "--runtime=60s", "--iodepth=256",
      "--verify_async=8", "--bs=4k-64k", "--norandommap", "--random_distribution=zipf:0.5",
      "--numjobs=16", "--rw=randrw", "--name=A", "--direct=1", "--name=B", "--direct=0"}

   workloadActive.Store(true)
   defer workloadActive.Store(false)

   err := launch(args...)
   if err != nil {
      return fmt.Errorf("workload: %w", err)
   }

   return nil
}
