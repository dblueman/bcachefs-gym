package main

import (
   "errors"
   "flag"
   "fmt"
   "os"
   "os/exec"
   "log/slog"
   "math/rand"
   "runtime/debug"
   "strconv"
   "strings"
   "sync"
   "time"
)

const (
   brdSize   = 512 << 20
   brdCount  = 16
   mountPath = "/mnt"
)

var (
   flagSeed  = flag.Int64("seed", -1, "random seed")
   blockDevs []string
   mountBusy sync.WaitGroup
)

func launch(args... string) error {
   cmd := exec.Command(args[0], args[1:]...)
   cmd.Stdout = os.Stdout
   cmd.Stderr = os.Stderr

   slog.Info("[" + strings.Join(args, " ") + "]")

   err := cmd.Run()
   if err != nil {
      slog.Error(err.Error())
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

func prob(vector *[]string, p float64, args... string) {
   if rand.Float64() > p {
      return
   }

   *vector = append(*vector, args...)
}

func pick[V any](options []V) V {
   i := rand.Intn(len(options))
   return options[i]
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

   go tunables()

   err = workload()
   if err != nil {
      return fmt.Errorf("cycle: %w", err)
   }

   mountBusy.Wait()

   return err
}

func _main() error {
   if os.Getuid() != 0 {
      return errors.New("please run as root")
   }

   info, _ := debug.ReadBuildInfo()
   sha := "unknown"
   mod := ""

   for _, elem := range info.Settings {
      switch elem.Key {
      case "vcs.revision":
         sha = elem.Value[:6]
      case "vcs.modified":
         if elem.Value == "true" {
            mod = "+"
         }
      }
   }

   slog.Info("build SHA", sha, mod)

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
   slog.SetDefault(
      slog.New(slog.NewTextHandler(os.Stderr, nil)),
   )

   flag.Usage = usage
   flag.Parse()

   if flag.NArg() != 0 {
      flag.Usage()
      os.Exit(2)
   }

   if *flagSeed != -1 {
      *flagSeed = time.Now().UnixNano()
      slog.Info("using seed %d", *flagSeed)
   }

   rand.Seed(*flagSeed)

   err := _main()
   if err != nil {
      fmt.Fprintln(os.Stderr, err.Error())
      os.Exit(1)
   }
}
