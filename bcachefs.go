package main

import (
   "math/rand"
   "fmt"
   "strconv"
   "strings"
)

var (
   compressors = []string{"none", "lz4", "gzip", "zstd"}
   dataHashes  = []string{"crc32c", "crc64", "siphash"}
   strHashes   = []string{"crc32c", "crc64", "xxhash"}
)

func pick[V any](options []V) V {
   i := rand.Intn(len(options))
   return options[i]
}

func format() error {
   args := []string{"bcachefs", "format"}

   prob(args, 0.1, "--data_replicas=" + strconv.Itoa(1 + rand.Intn(len(blockDevs))))
   prob(args, 0.1, "--data_replicas_required=" + strconv.Itoa(1 + rand.Intn(len(blockDevs))))
   prob(args, 0.1, "--data_checksum=" + pick(dataHashes))

   prob(args, 0.1, "--metadata_replicas=" + strconv.Itoa(1 + rand.Intn(len(blockDevs))))
   prob(args, 0.1, "--metadata_replicas_required=" + strconv.Itoa(1 + rand.Intn(len(blockDevs))))
   prob(args, 0.1, "--metadata_checksum=" + pick(dataHashes))

   // prob(args, 0.1, "--erasure_code") // skip for now as "not stable"
   prob(args, 0.1, "--inodes_32bit")
   prob(args, 0.1, "--shard_inode_numbers")
   prob(args, 0.1, "--inodes_use_key_cache")
   prob(args, 0.1, "--gc_reserve_percent=" + strconv.Itoa(0 + rand.Intn(100)))
   prob(args, 0.1, "--root_reserve_percent=" + strconv.Itoa(0 + rand.Intn(100)))
   prob(args, 0.1, "--wide_macs")
   prob(args, 0.1, "--acl")
   prob(args, 0.1, "--usrquota")
   prob(args, 0.1, "--grpquota")
   prob(args, 0.1, "--prjquota")
   prob(args, 0.1, "--journal_transaction_names")
   prob(args, 0.1, "--nocow")
   prob(args, 0.1, "--acl")
   prob(args, 0.1, "--encrypted")
   prob(args, 0.1, "--compression=" + pick(compressors))
   prob(args, 0.1, "--str_hash=" + pick(strHashes))
   prob(args, 0.1, "--block_size=4096")
   prob(args, 0.1, "--btree_node_size=262144")
   prob(args, 0.1, "--no_passphrase")
   prob(args, 0.1, "--fs_label=test")
   prob(args, 0.1, "--shard_inode_numbers")
   prob(args, 0.1, "--background_compression=" + pick(compressors))

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
   return nil
}
