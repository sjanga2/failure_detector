package main

import (
	"cs425_g12/common"
	"cs425_g12/gossip"
	"cs425_g12/hydfs_utils"
	"fmt"
	"os"
	"strconv"
	"time"
)

// map from machine number to ip address
var ip_map = map[string]string{
	"01": "172.22.94.224",
	"02": "172.22.154.39",
	"03": "172.22.158.39",
	"04": "172.22.94.225",
	"05": "172.22.154.40",
	"06": "172.22.158.40",
	"07": "172.22.94.226",
	"08": "172.22.154.41",
	"09": "172.22.158.41",
	"10": "172.22.94.227",
}

// times for passing in to checkers
var Tsus = 2 * time.Second
var Tfail = 3 * time.Second
var Tclean = 6 * time.Second
var Tgossip = 200 * time.Millisecond
var Tping = 500 * time.Millisecond
var Tsuscheck = 500 * time.Millisecond
var Tfailcheck = 500 * time.Millisecond

var introducer_ip = "172.22.94.224"

func main() {
	machineNo := os.Args[1]
	protocol := os.Args[2]
	runSus := os.Args[3]
	DropRate, err := strconv.ParseFloat(os.Args[4], 64)

	hydfs_utils.InitHyDFSDir()

	// initialize gossip logger and hydfs logger
	common.InitializeLogger("machine" + machineNo)
	hydfs_utils.InitializeLogger("machine" + machineNo)

	if err != nil {
		common.Logger.Printf("Error converting drop rate to float; defaulting to 0.0 drop rate")
		DropRate = 0.0
	}

	common.SetSelf(common.NewMachineId(ip_map[machineNo], common.GlobalPort, time.Now()))

	introducer := common.MachineId{
		Ip:   introducer_ip,
		Port: common.GlobalPort,
	}

	list := common.NewMembershipList()
	common.Introducer = introducer

	if common.GetSelf().Ip == introducer_ip {
		fmt.Println("this is the introducer")
		introMember := common.NewMember(common.GetSelf())
		fmt.Printf("introducer member: %+v\n", introMember)
		list.Insert(introMember)
		common.IsMemberInGroup = true
	} else {
		fmt.Println("this is a normal machine")
		// if not introducer,  send request to introducer to join
		joinReqSuccess := gossip.RequestJoin(introducer, list)
		if !joinReqSuccess {
			fmt.Println("Join request failed, exiting...")
			return
		}
		common.IsMemberInGroup = true
	}

	if runSus == "withSus" {
		common.SetSuspicionMode(true)
	} else if runSus == "withNoSus" {
		common.SetSuspicionMode(false)
	}

	if protocol == "pingack" {
		common.SetProtocolMode(true)
	} else if protocol == "gossip" {
		common.SetProtocolMode(false)
	}

	// GOSSIP GOROUTINES
	list.StartChecker(Tsus, Tfail, Tclean, Tsuscheck, Tfailcheck)
	gossip.GossipListener(list, DropRate)
	gossip.StartProtocol(list, Tgossip, Tping, Tfail)

	// HYDFS GOROUTINES

	go func() {
		for {
			var command, protocolMode, susMode string
			fmt.Scanln(&command, &protocolMode, &susMode)

			switch command {
			// FAILURE DETECTOR COMMANDS
			case "switch":
				if protocolMode == "gossip" && susMode == "withNoSus" {
					common.SetSuspicionMode(false)
					common.SetProtocolMode(false)
					common.Logger.Println("Switched to gossip withNoSus mode.")
				} else if protocolMode == "gossip" && susMode == "withSus" {
					common.SetSuspicionMode(true)
					common.SetProtocolMode(false)
					common.Logger.Println("Switched to gossip withSus mode.")
				} else if protocolMode == "pingack" && susMode == "withNoSus" {
					common.SetSuspicionMode(false)
					common.SetProtocolMode(true)
					common.Logger.Println("Switched to pingack withNoSus mode.")
				} else if protocolMode == "pingack" && susMode == "withSus" {
					common.SetSuspicionMode(true)
					common.SetProtocolMode(true)
					common.Logger.Println("Switched to pingack withSus mode.")
				} else {
					common.Logger.Printf("Invalid switch parameters: %s %s", protocolMode, susMode)
				}
			case "list_mem":
				membersPrint := list.GetEntireList()
				fmt.Println("Membership list : ")
				for _, m := range membersPrint {
					fmt.Println("  ", m)
				}
				common.Logger.Printf("Called getEntireList")
			case "list_self":
				fmt.Println("Self ID:", common.GetSelf())
				common.Logger.Printf("Called getSelf")
			case "leave":
				self := common.GetSelf()
				if member := list.GetMember(self); member != nil {
					member.SuspicionState = common.StateFailed
				}
				common.IsMemberInGroup = false
				fmt.Println("Left the group voluntarily")
			case "join":
				if common.IsMemberInGroup {
					fmt.Println("already in the group")
				} else {
					joined := gossip.RequestJoin(common.Introducer, list)
					if joined {
						common.IsMemberInGroup = true
						fmt.Println("Joined the group")
					} else {
						fmt.Println("Failed to Join")
					}
				}
			case "display_suspects":
				suspectedMachines := []common.Member{}
				for _, machine := range list.GetEntireList() {
					if machine.SuspicionState == common.StateSuspicious {
						suspectedMachines = append(suspectedMachines, machine)
					}
				}
				if len(suspectedMachines) == 0 {
					fmt.Println("No suspected machines")
				} else {
					fmt.Println("Suspected machines: ")
					for _, suspect := range suspectedMachines {
						fmt.Println(" ", suspect)
					}
				}
			case "display_protocol":
				var protocol string
				var sus string
				if common.GetProtocolMode() {
					protocol = "ping"
				} else {
					protocol = "gossip"
				}
				if common.GetSuspicionMode() {
					sus = "suspect"
				} else {
					sus = "nosuspect"
				}
				fmt.Printf("<%s, %s>\n", protocol, sus)
			case "start_exp":
				gossip.IsExperimentRunning.Store(true)
				go gossip.LogExperiments()
				fmt.Println("Experiment started, logging bandwidth stats.")
			case "stop_exp":
				gossip.IsExperimentRunning.Store(false)
				fmt.Println("Experiment stopped.")

			// HYDFS COMMANDS

			default:
				common.Logger.Printf("Unknown command: %s %s %s", command, protocolMode, susMode)
			}
		}
	}()

	fmt.Println("gossip/pingack started; ur life is ruined")
	select {} // block forever
}
