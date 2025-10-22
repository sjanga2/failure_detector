package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/exec"
)

// rpc service for running shell commands
type CommandExecutor struct{}

func (c *CommandExecutor) Run(args []string, reply *string) error {
	// grep command
	command := args[0]
	// mode: demo or unit_test
	mode := args[1]
	// machine number from server args
	machineNo := os.Args[1]

	var logFilename string
	if mode == "unit_test" {
		logFilename = fmt.Sprintf("/home/shared/machine%s.log", machineNo)
	} else if mode == "demo" {
		logFilename = fmt.Sprintf("../vm%s.log", machineNo)
	}
	// execute grep command passed in
	cmd := exec.Command("sh", "-c", command+" "+logFilename)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// add print message for when grep command will fail
		fmt.Println("Grep command failed:", err)
		return err
	}
	*reply = string(out)
	return nil
}

func main() {
	// new CommandExecutor instance
	executor := new(CommandExecutor)

	// registering executor for RPC
	rpc.Register(executor)

	// listening to tcp connections on this port
	listener, err := net.Listen("tcp", ":"+"1234")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	// closes listener when func ends (but never called because func never exits)
	defer listener.Close()
	fmt.Println("Server listening on (ipaddress for machines?)port ", 1234)

	// accepting any no of incoming connections - handles multiple client connections
	for {
		connection, err := listener.Accept()
		if err != nil {
			continue
		}
		// serve the current connection through rpc
		go rpc.ServeConn(connection)
	}
}
