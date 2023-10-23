package main

import (
   "errors"
   "flag"
   "fmt"
   "os"
   "os/exec"
   "log/slog"
   "math/rand"
   "strconv"
   "strings"
   "time"
)

const (
   brdSize   = 512 << 20
   brdCount  = 16
   mountPath = "/mnt"
)

var (
   flagSeed  = flag.Int64("seed", -1, "random seed")
   log       = slog.New(slog.NewTextHandler(os.Stderr, nil))
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

func prob(vector []string, p float32, args... string) {
   if rand.Float32() < p {
      return
   }

   vector = append(vector, args...)
}

func cycle() error {
   err := launch("modprobe", "brd",
      "rd_size=" + strconv.Itoa(brdSize),
      "rd_nr=" + strconv.Itoa(brdCount),
   )
   if err != nil {
      return fmt.Errorf("cycle: %w", err)
   }

   defer func() {
      // remove any existing
      err = launch("rmmod", "brd")
   }()

   blockDevs = []string{}

   for i := 0; i < 1 + rand.Intn(brdCount - 1); i++ {
      blockDevs = append(blockDevs,
         fmt.Sprintf("/dev/ram%d", i),
      )
   }

   err = format()
   if err != nil {
      return fmt.Errorf("cycle: %w", err)
   }

   err = mount()
   if err != nil {
      return fmt.Errorf("cycle: %w", err)
   }

   defer func() {
      err = unmount()
   }()

   err = workload()
   if err != nil {
      return fmt.Errorf("cycle: %w", err)
   }

   return err
}

func _main() error {
   if os.Getuid() != 0 {
      return errors.New("please run as root")
   }

   for {
      err := cycle()
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

   if *flagSeed != -1 {
      *flagSeed = time.Now().UnixNano()
      log.Info("using seed %d", *flagSeed)
   }

   rand.Seed(*flagSeed)

   err := _main()
   if err != nil {
      fmt.Fprintln(os.Stderr, err.Error())
      os.Exit(1)
   }
}
