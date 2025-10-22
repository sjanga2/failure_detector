# CS425_G12

## General running command:
`go run main.go <machineNum> <protocol> <susmode> <dropRate>`

- machineNum: 01, 02, ... 10
- protocol: `gossip` or `pingack`
- susmode: `withSus` or `withNoSus`
- dropRate: decimal between 0 to 1, 0 denotes no messages dropped.

## Running the introducer:

Currently, the introducer is hardcoded to machine 1201. So, ssh into machine 1201 and do the following:
1. `cd ~/cs425_g12`
2. `go run main.go 01 <protocol> <susmode> <dropRate>`

The introducer must always be the first machine to run.

## Running all other machines:

Once the introducer is ready, add the other machines using the same steps as above.

## Other guidelines:

1. Logs are saved in `/home/shared/machine<machineNum>.log`.
2. Anytime a machine is marked as suspicious or failed, it is printed to stdout.
3. The following commands are available to interface with the failure detector:
    - list_mem: list the membership list
    - list_self: list self’s id
    - join: join the group (it is ok for this command to be implicitly executed when the process starts, or you could implement the command explicitly)
    - leave: voluntarily leave the group (different from a failure)
    - display_suspects: List suspected nodes.
    - switch {gossip, ping}, {suspect, nosuspect}: it switches the current mechanism to gossip/ping (whichever is first parameter), and without or without suspicion (second parameter).