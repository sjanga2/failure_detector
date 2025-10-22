package gossip

import (
	"cs425_g12/common"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sync/atomic"
	"time"
)

// ip address of all machines
var ip_addresses = []string{
	"172.22.94.224",
	"172.22.154.39",
	"172.22.158.39",
	"172.22.94.225",
	"172.22.154.40",
	"172.22.158.40",
	"172.22.94.226",
	"172.22.154.41",
	"172.22.158.41",
	"172.22.94.227",
}

// port for udp communication
const gossipPort = 5051

var globalConn net.PacketConn

// varibles for measuring bandwidth
var experimentBytesSent uint64
var experimentBytesRecv uint64
var IsExperimentRunning atomic.Bool

func recordSend(size int) {
	if IsExperimentRunning.Load() {
		atomic.AddUint64(&experimentBytesSent, uint64(size))
	}
}

func recordRecv(size int) {
	if IsExperimentRunning.Load() {
		atomic.AddUint64(&experimentBytesRecv, uint64(size))
	}
}

func LogExperiments() {
	file, _ := os.OpenFile("experiment_bandwidth.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer file.Close()
	logger := log.New(file, "", log.LstdFlags)

	// recording for every 2 mintues
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for IsExperimentRunning.Load() {
		<-ticker.C

		sent := atomic.SwapUint64(&experimentBytesSent, 0)
		recv := atomic.SwapUint64(&experimentBytesRecv, 0)

		avgSentSec := float64(sent) / (2 * 60)
		avgRecvSec := float64(recv) / (2 * 60)

		logger.Printf("Bytes avg sent ==%.2f B/s, recv=%.2f B/s, total=%.2f B/s\n", avgSentSec, avgRecvSec, avgRecvSec+avgSentSec)
	}
}

// this is the info sent over to other machines
type GossipInfo struct {
	MemberSummary []common.Member  // this is the output of the getter GetEntireList, don't confuse this with the MembershipList struct which contains the mutex and a map of members
	Sender        common.MachineId // this is to identify the sender, not sure if needed
}

type MessageType struct {
	Type string // "join", "gossip", "updatedList"
	Data json.RawMessage
}

func SendGossip(list *common.MembershipList, Tgossip time.Duration) {
	// increment own heartbeat counter
	list.IncrementHeartbeat()

	self := common.GetSelf()

	if len(list.GetEntireList()) == 1 && list.GetEntireList()[0].MachineId.Ip == self.Ip {
		return // only self in the list, skip gossip
	}

	// pick a random index from the membership list to gossip to, but don't choose urself
	var target string
	for {
		currList := list.GetUniqueMembers()
		if len(currList) == 0 {
			common.Logger.Println("Membership list is empty, skipping gossip")
			time.Sleep(Tgossip)
			continue
		}
		target = currList[rand.Intn(len(currList))].MachineId.Ip
		if target != self.Ip {
			//cannot choose self as target
			break
		}
	}

	common.Logger.Printf("Chose target %s for gossip\n", target)

	// get the entire list to send
	currList := list.GetEntireList()
	infoToSend := GossipInfo{
		MemberSummary: currList,
		Sender:        self,
	}

	// wrap the gossip info in a MessageType
	msg := MessageType{
		Type: "gossip",
		Data: func() json.RawMessage {
			data, err := json.Marshal(infoToSend)
			if err != nil {
				fmt.Println("Error occurred during marshaling: ", err)
				return nil
			}
			return data
		}(),
	}

	// "marshal" the message to send
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error occurred during marshaling: ", err)
		return
	}

	common.Logger.Printf("Data to be sent: %s\n", string(data))
	recordSend(len(data))

	// randomly chosen target with hardcoded port, not sure if single port is ok
	addr := fmt.Sprintf("%s:%d", target, gossipPort)

	readAddr := &net.UDPAddr{IP: net.ParseIP(target), Port: gossipPort}
	_, err = globalConn.WriteTo(data, readAddr)
	if err != nil {
		fmt.Println("error sending ping: ", err)
	}

	// log this gossip event
	common.Logger.Printf("Gossiped to %s with %d members\n", addr, len(currList))
}

// listen for gossip
func GossipListener(list *common.MembershipList, DropRate float64) {
	self := common.GetSelf()
	addr := fmt.Sprintf("%s:%d", self.Ip, self.Port)

	var err error
	// udp listener setup
	globalConn, err = net.ListenPacket("udp", addr)
	if err != nil {
		fmt.Println("Error setting up UDP listener: ", err)
		return
	}

	go func() {
		// defer globalConn.Close()
		buffer := make([]byte, 65535) // configured with random large size for now

		for {
			bytesRead, from, err := globalConn.ReadFrom(buffer)

			common.Logger.Printf("Received message from %s: %s\n", from.String(), string(buffer[:bytesRead]))
			if rand.Float64() < DropRate {
				common.Logger.Print("Dropping the message.")
				continue
			}

			if err != nil {
				fmt.Println("Error reading from UDP: ", err)
				continue
			}
			recordRecv(bytesRead)

			// decode the msg type
			var msgType MessageType
			err = json.Unmarshal(buffer[:bytesRead], &msgType)
			if err != nil {
				fmt.Println("Error unmarshaling message type: ", err)
				continue
			}
			//join message
			if msgType.Type == "join" {
				var newMember common.Member
				err = json.Unmarshal(msgType.Data, &newMember)
				if err != nil {
					fmt.Println("Error unmarshaling new member: ", err)
					continue
				}

				// insert the new member into the membership list
				list.Insert(newMember)
				common.Logger.Printf("New member joined: %+v\n", newMember)

				// send list back to new joiner
				members := list.GetEntireList()
				reply := MessageType{
					Type: "updatedList",
					Data: func() json.RawMessage {
						data, err := json.Marshal(members)
						if err != nil {
							fmt.Println("Error marshaling member list: ", err)
							return nil
						}
						return data
					}(),
				}
				globalConn.WriteTo([]byte(func() string {
					data, err := json.Marshal(reply)
					if err != nil {
						fmt.Println("Error marshaling reply: ", err)
						return ""
					}
					recordSend(len(data))
					return string(data)
				}()), from)

				// handling gossip message
			} else if msgType.Type == "gossip" {
				// unmarshal the Data field into receivedInfo
				var receivedInfo GossipInfo
				err = json.Unmarshal(msgType.Data, &receivedInfo)
				if err != nil {
					fmt.Println("Error unmarshaling gossip info: ", err)
					continue
				}

				self = common.GetSelf()

				// merging the incoming membership list into own
				MergeGossip(list, receivedInfo.MemberSummary, self)

				// common.Logger.Printf("Received gossip from %s with %d members\n", from.String(), len(receivedInfo.MemberSummary))
				common.Logger.Printf("Received gossip from %s\n", receivedInfo.Sender)
				for _, m := range receivedInfo.MemberSummary {
					common.Logger.Printf("  %s\n", m)
				}
				// handling ping messages
			} else if msgType.Type == "ping" {
				var ping Ping
				json.Unmarshal(msgType.Data, &ping)
				// merging received membership list to own (piggpy back)
				MergePingAck(list, ping.MemberSummary)

				ack := Ack{
					Sender:        self,
					MemberSummary: list.GetEntireList(),
				}

				reply := MessageType{Type: "ack", Data: helperMarshal(ack)}
				out := helperMarshal(reply)
				// send ack
				globalConn.WriteTo(out, from)
				recordSend(len(out))
			} else if msgType.Type == "ack" {
				var ack Ack
				json.Unmarshal(msgType.Data, &ack)

				// merging received membership list to own (piggpy back)
				MergePingAck(list, ack.MemberSummary)

				// merging logic should handle every change, don't need to explicitly modify anything
				// if ackMachine := list.GetMember(ack.Sender); ackMachine != nil {
				// ackMachine.SuspicionState = common.StateAlive
				// ackMachine.TimeLocal = time.Now()
				// shouldnt update from fail to alive
				// } else {
				//needs to rejoin - handle self failure
				// }

				ackMutex.Lock()
				ackReceived = true
				ackMutex.Unlock()
			}
		}
	}()
}

// called by the main function during init when a new machine needs to be introduced to the group
func RequestJoin(introducer common.MachineId, list *common.MembershipList) bool {
	self := common.GetSelf()
	joinMember := common.NewMember(self)

	// sending out a join message of self
	msgType := MessageType{
		Type: "join",
		Data: func() json.RawMessage {
			data, err := json.Marshal(joinMember)
			if err != nil {
				fmt.Println("Error marshaling joinMember: ", err)
				return nil
			}
			return data
		}(),
	}

	addr := fmt.Sprintf("%s:%d", introducer.Ip, introducer.Port)

	//dialing on the introducer
	conn, err := net.Dial("udp", addr)
	if err != nil {
		fmt.Println("Error dialing introducer: ", err)
		return false
	}
	defer conn.Close()

	data, err := json.Marshal(msgType)
	if err != nil {
		fmt.Println("Error marshaling join message: ", err)
		return false
	}

	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Error sending join request: ", err)
		return false
	}
	recordSend(len(data))

	buffer := make([]byte, 65535)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // wait for 5 seconds max, if not responds, introducer is prob dead
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading join response: ", err)
		return false
	}

	// decode the msg type
	var introducerMsg MessageType
	err = json.Unmarshal(buffer[:bytesRead], &introducerMsg)
	if err != nil {
		fmt.Println("Error unmarshaling join response: ", err)
		return false
	}

	if introducerMsg.Type == "updatedList" {
		var members []common.Member
		err = json.Unmarshal(introducerMsg.Data, &members)
		if err != nil {
			fmt.Println("Error unmarshaling member list: ", err)
			return false
		}

		for _, member := range members {
			if member.SuspicionState != common.StateFailed {
				// only inserting non failed members
				list.Insert(member)
			}
		}

		for _, m := range list.GetSortedRing() {
			fmt.Printf("   Sorted Ring Member: %s\n", *m)
		}

		fmt.Printf("    Predecessor: %s\n", list.FindPredecessor(list.GetMember(common.GetSelf()).RingId))
		fmt.Printf("    Successor: %s\n", list.FindSuccessor(list.GetMember(common.GetSelf()).RingId))
		fmt.Printf("    Two successors: %s\n", list.GetSuccessorNodes(list.GetMember(common.GetSelf()).RingId, 2))

		common.Logger.Printf("Joined group, got %d members\n", len(members))
		return true
	}

	return false
}

func handleSelfFailure(list *common.MembershipList) {

	if !common.IsMemberInGroup {
		// not a member of the group, no need to rejoin
		return
	}

	// clear the entire membership list
	self := common.GetSelf()

	selfNewVersion := common.NewMachineId(self.Ip, self.Port, time.Now())

	if (self.Ip == common.Introducer.Ip) && (self.Port == common.Introducer.Port) {
		// no need to delete entire list
		// inserting new self
		list.Delete(self)
		common.SetSelf(selfNewVersion)
		list.Insert(common.NewMember(selfNewVersion))
		common.Logger.Printf("I am the introducer.")
	} else {
		// delete entire list and request to rejoin
		list.DeleteEntireList()
		common.SetSelf(selfNewVersion)
		RequestJoin(common.Introducer, list)
		common.Logger.Printf("I am not the introducer. The introducer is: %+v", common.Introducer)
	}

	// common.SetSelf(selfNewVersion)

	// RequestJoin(common.Introducer, list)

	common.Logger.Printf("Handled self failure, new MachineId: %+v\n", selfNewVersion)

}
