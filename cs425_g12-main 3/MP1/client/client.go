package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"
	"strconv"
)

func main() {
	// mode: demo or unit_test
	mode := os.Args[1]

	// list of all the ip address of vms
	ip_addresses := []string{"172.22.94.224", "172.22.154.39", "172.22.158.39", "172.22.94.225", "172.22.154.40",
		"172.22.158.40", "172.22.94.226", "172.22.154.41", "172.22.158.41", "172.22.94.227"}

	// store servers in a list to dial, call and close. remember machine numbers for printing later
	servers := make([]*rpc.Client, 0, len(ip_addresses))
	machine_nos := make([]int, 0, len(ip_addresses))

	// loop through ports and dial
	for i, ip := range ip_addresses {
		address := fmt.Sprintf("%s:1234", ip)

		server, err := rpc.Dial("tcp", address)
		if err != nil {
			// machine is not available, continue to next port
			fmt.Printf("Failed to connect to %s: \n", address)
			continue
		}
		// record the successfully connected server and schedule to close later. also maintain a mapping to original machine number
		servers = append(servers, server)
		machine_nos = append(machine_nos, i+1)
		defer server.Close()
		// fmt.Printf("Machine no: %d \n", i)
	}

	fmt.Println("Enter grep commands (type 'exit' to quit):")

	// read in the commands from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		if command == "exit" {
			break
		}

		// check if the command is grep -c
		isGrepCount := false
		if command[:7] == "grep -c" {
			isGrepCount = true
		}
		// total sum of matches across all files
		sum := 0

		// arguments for the server
		args := []string{command, mode}

		// loop through all the servers and call the Run method
		// fmt.Printf("Size of servers array: %d", len(servers))
		for i, server := range servers {
			var reply string
			// fmt.Printf("I am accessing machine nos idx %d \n",i)
			err := server.Call("CommandExecutor.Run", args, &reply)
			if err != nil {
				// machine did not respond, print error and continue
				fmt.Printf("Reply from machine %d: error %s\n", machine_nos[i], err.Error())
				continue
			}
			fmt.Printf("Reply from machine %d: %s\n", machine_nos[i], reply)
			// if grep -c, maintain the sum of matches
			if isGrepCount && (err == nil) {
				num, convErr := strconv.Atoi(reply[:len(reply)-1])
				if (convErr == nil) {
					sum += num
				}
			}
		}
		// print total matches if grep -c
		if isGrepCount {
			fmt.Printf("Total matches across all files: %d\n", sum)
		}
	}
}
