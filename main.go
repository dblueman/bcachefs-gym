package main

import (
   "flag"
   "fmt"
   "os"
   "os/exec"
   "log/slog"
   "strconv"
   "strings"
)

const (
   brdSize   = 512 << 20
   brdCount  = 16
   mountPath = "/mnt"
)

var (
   log = slog.New(slog.NewTextHandler(os.Stderr, nil))
   blockDevs []string
)

func launch(args... string) error {
   cmd := exec.Command(args[0], args[1:]...)
   cmd.Stdout = os.Stdout
   cmd.Stderr = os.Stderr

   log.Info("[" + strings.Join(args, " ") + "]")

   err := cmd.Run()
   if err != nil {
      log.Error(err.Error())
      return fmt.Errorf("launch: %w", err)
   }

   return nil
}

func unmount() error {
   err := launch("umount", mountPath)
   if err != nil {
      return fmt.Errorf("unmoun: %w", err)
   }

   return nil
}

func _main() error {
   // remove any existing
   _ = launch("rmmod", "brd")

   err := launch("modprobe", "brd",
      "rd_size=" + strconv.Itoa(brdSize),
      "rd_Count=" + strconv.Itoa(brdCount),
   )
   if err != nil {
      return err
   }

   for i := 0; i < brdCount; i++ {
      blockDevs = append(blockDevs,
         fmt.Sprintf("/dev/ram%d", i),
      )
   }

   for {
      err = create()
      if err != nil {
         return err
      }

      err = mount()
      if err != nil {
         return err
      }

      err = workload()
      if err != nil {
         return err
      }

      err = unmount()
      if err != nil {
         return err
      }
   }

   return nil
}

func usage() {
   fmt.Fprintln(os.Stderr, "usage: bcachefs-gym")
   flag.PrintDefaults()
}

func main() {
   flag.Usage = usage
   flag.Parse()

   if flag.NArg() != 0 {
      flag.Usage()
      os.Exit(2)
   }

   err := _main()
   if err != nil {
      fmt.Fprintln(os.Stderr, err.Error())
      os.Exit(1)
   }
}
