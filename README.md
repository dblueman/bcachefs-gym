# BcacheFS-gym
BcacheFS-gym is tool to exercise the Linux BcacheFS filesystem, without using local storage.

## How to use it
1. First build it using Go 1.20 or newer:
    ```
    $ git clone https://github.com/dblueman/bcachefs-gym
    $ cd bcachefs-gym
    $ go install
    ```
1. Launch it as root:
    ```
    $ sudo ./bcachefs-gym
    ```

## How it works
1. BcacheFS-gym loads the brd block-ramdisk kernel module to create in-memory block devices
1. It runs the *bcachefs* userspace utility to format a number of block ram devices, randomising format options with constraints
1. The block devices are mounted using randomised mount options, constrained
1. A short fio workload is launched with direct and pagecache IO
1. Background *bcachefs* maintenance commands are launched at regular intervals
1. Once the workload completes, the filesystem is unmounted and the process repeats
1. At any point, if there are errors (non-zero exit code), execution is stopped for analysis
