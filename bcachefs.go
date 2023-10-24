package main

import (
   "math/rand"
   "fmt"
   "strconv"
   "strings"
   "time"
)

var (
   compressors = []string{"none", "lz4", "gzip", "zstd"}
   checksums   = []string{"none", "crc32c", "crc64"}
   strHashes   = []string{"crc32c", "crc64", "siphash"}
)

func format() error {
   args := []string{"bcachefs", "format"}
   p := 0.2
   maxReplicas := min(len(blockDevs), 3)

   // FIXME add later when "unable to write journal to sufficient devices" error resolved
   // prob(&args, p, "--data_replicas_required=" + strconv.Itoa(1 + rand.Intn(maxReplicas)))
   // prob(&args, p, "--metadata_replicas_required=" + strconv.Itoa(1 + rand.Intn(maxReplicas)))

   prob(&args, p, "--data_replicas=" + strconv.Itoa(1 + rand.Intn(maxReplicas)))
   prob(&args, p, "--data_checksum=" + pick(checksums))
   prob(&args, p, "--metadata_replicas=" + strconv.Itoa(1 + rand.Intn(maxReplicas)))
   prob(&args, p, "--metadata_checksum=" + pick(checksums))

   // prob(&args, p, "--erasure_code") // FIXME add later when stable
   prob(&args, p, "--inodes_32bit")
   prob(&args, p, "--shard_inode_numbers")
   prob(&args, p, "--inodes_use_key_cache")
   prob(&args, p, "--gc_reserve_percent=" + strconv.Itoa(5 + rand.Intn(20 + 1 - 5))) // 5-20
   prob(&args, p, "--root_reserve_percent=" + strconv.Itoa(0 + rand.Intn(100)))
   prob(&args, p, "--wide_macs")
   prob(&args, p, "--acl")
   prob(&args, p, "--usrquota")
   prob(&args, p, "--grpquota")
   prob(&args, p, "--prjquota")
   prob(&args, p, "--journal_transaction_names")
   prob(&args, p, "--nocow")
   prob(&args, p, "--acl")
   prob(&args, p, "--encrypted", "--no_passphrase")
   prob(&args, p, "--compression=" + pick(compressors))
   prob(&args, p, "--str_hash=" + pick(strHashes))
   prob(&args, p, "--block_size=4096")
   prob(&args, p, "--btree_node_size=262144")
   prob(&args, p, "--fs_label=test")
   prob(&args, p, "--shard_inode_numbers")
   prob(&args, p, "--background_compression=" + pick(compressors))

   // TODO
   // --durability=
   // --foreground_target=
   // --metadata_target=
   // --promote_target=

   args = append(args, blockDevs...)

   err := launch(args...)
   if err != nil {
      return fmt.Errorf("create: %w", err)
   }

   return nil
}

func mount() error {
   args := []string{"mount", "-t", "bcachefs"}

   // options: degraded very_degraded verbose fsck fix_errors ratelimit_errors norecovery version_upgrade discard

   args = append(args, strings.Join(blockDevs, ":"))
   args = append(args, mountPath)

   err := launch(args...)
   if err != nil {
      return fmt.Errorf("create: %w", err)
   }

   return nil
}

func tunables() error {
   // parameters:
   // journal_flush_delay
   // journal_flush_disabled

   // bcachefs options:
   // device add/remove/online/offline/evacuate/set-state
   // data rereplicate
   // subvolume create/snapshot/destroy
   // show-super [-l] [-f all]
   // list
   // list_journal
   // fs usage
   // attr ...

   // runtime options via /sys/fs/bcachefs/<uuid>/options/

   mountBusy.Add(1)

   commands := []string{
      "add", "remove", "online", "offline",
      "evacuate", "set-state", "resize", "resize-journal",
   }

   time.Sleep(10 * time.Second)

   for workloadActive.Load() {
      args := []string{"bcachefs", "device", pick(commands)}

      if args[2] == "add" {
         args = append(args, mountPath)
      }

      args = append(args, pick(blockDevs))

      if args[2] == "resize-journal" {
         args = append(args, strconv.Itoa(rand.Intn(brdSize)))
      }

      launch(args...)

      time.Sleep(3 * time.Second)
   }

   mountBusy.Done()

   return nil
}
