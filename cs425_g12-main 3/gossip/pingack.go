package gossip

import (
	"cs425_g12/common"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

// ping messages data
type Ping struct {
	Sender        common.MachineId
	MemberSummary []common.Member
}

// ack messages data
type Ack struct {
	Sender        common.MachineId
	MemberSummary []common.Member
}

// tracker for if ack has been received
var (
	ackReceived bool
	ackMutex    sync.Mutex
)

func StartPinging(list *common.MembershipList, Tfail time.Duration) {

	list.IncrementHeartbeat()

	self := common.GetSelf()
	members := list.GetUniqueMembers()

	if len(members) == 1 && members[0].MachineId.Ip == self.Ip {
		return // only self in the list, skip the pinging
	}

	var target common.MachineId

	for {
		//only choosing non self target
		chosen := members[rand.Intn(len(members))]
		if chosen.MachineId.Ip != self.Ip {
			target = chosen.MachineId
			break
		}
	}

	// handling ping and waiting for ack with a time out
	PingAndWait(target, list, Tfail)
}

func PingAndWait(target common.MachineId, list *common.MembershipList, Tfail time.Duration) {
	self := common.GetSelf()

	// ack tracker reset to false
	ackMutex.Lock()
	ackReceived = false
	ackMutex.Unlock()

	ping := Ping{
		Sender:        self,
		MemberSummary: list.GetEntireList(),
	}

	pingMessage := MessageType{Type: "ping", Data: helperMarshal(ping)}

	data, _ := json.Marshal(pingMessage)

	common.Logger.Printf("data to be sent: %s\n", string(data))

	addr := fmt.Sprintf("%s:%d", target.Ip, target.Port)

	readAddr := &net.UDPAddr{IP: net.ParseIP(target.Ip), Port: int(target.Port)}
	// write the data to address
	_, err := globalConn.WriteTo(data, readAddr)
	if err != nil {
		fmt.Println("error sending ping: ", err)
	}
	recordSend(len(data))

	common.Logger.Printf("sent ping to %s with %d members piggbacked ", addr, len(ping.MemberSummary))
	// recordSend()
	// set ack receive timeout
	ackTimeout := time.Now().Add(Tfail)

	// wait till ack as long as before timeout
	for time.Now().Before(ackTimeout) {
		time.Sleep(5 * time.Millisecond)

		ackMutex.Lock()
		if ackReceived {
			ackMutex.Unlock()
			common.Logger.Printf("Ack received from %s", target)
			// ack received, no need to handle further
			return
		}
		ackMutex.Unlock()
	}

	// havent received ack and timer has expired
	// common.Logger.Printf("Have not received ack. Marking machine %s as failed.", target)
	// if failedTargetEntry := list.GetMember(target); failedTargetEntry != nil {
	// 	failedTargetEntry.SuspicionState = common.StateFailed
	// }

	// target did not respond marking as failed or sus based on mode
	if failedTargetEntry := list.GetMember((target)); failedTargetEntry != nil {
		if common.GetSuspicionMode() {
			if failedTargetEntry.SuspicionState == common.StateAlive {
				failedTargetEntry.SuspicionState = common.StateSuspicious
				fmt.Printf("Marked %s as suspicious (no ack)", target)
				common.Logger.Printf("Marked %s as suspicious (no ack)", target)
			}
		} else {
			fmt.Printf("[%s] Member %+v marked as Failed (from gossip)\n", time.Now().Format("15:04:05.000"), failedTargetEntry.MachineId)
			failedTargetEntry.SuspicionState = common.StateFailed
			common.Logger.Printf("DropSearch Have not received ack. Marking machine %s as failed.", target)
		}
	}
}

func MergePingAck(list *common.MembershipList, received []common.Member) {
	common.Logger.Printf("Merge Ping Ack function entered. Received gossip: %+v\n", received)
	self := common.GetSelf()

	now := time.Now()

	// merging the received with current
	for _, receivedMember := range received {
		currentListMember := list.GetMember(receivedMember.MachineId)


		if currentListMember == nil {
			// new member because current does not have it
			if receivedMember.SuspicionState != common.StateFailed {
				// insert if not failed 
				receivedMember.TimeLocal = now
				list.Insert(receivedMember)
				common.Logger.Printf("Added new member: %+v\n", receivedMember)
			}
			continue
		}

		if receivedMember.SuspicionState == common.StateFailed {
			// failure overrides everything
			if currentListMember.SuspicionState != common.StateFailed {
				fmt.Printf("[%s] Member %+v marked as Failed (from pingack)\n", time.Now().Format("15:04:05.000"), currentListMember.MachineId)
				common.Logger.Printf("DropSearch First notification of failure for member: %+v\n", currentListMember)
			}
			// updating everything except the time
			currentListMember.HeartbeatCounter = receivedMember.HeartbeatCounter
			currentListMember.SuspicionState = common.StateFailed
			currentListMember.IncarnationNumber = receivedMember.IncarnationNumber
			common.Logger.Printf("Updated member to Failed state: %+v\n", currentListMember)
			if currentListMember.MachineId == self {
				// handle failure
				handleSelfFailure(list)
			}
			continue
		}

		if currentListMember.SuspicionState == common.StateFailed {
			continue // do nothing, we have the newest info
		}

		// higher inc number always takes priority
		if receivedMember.IncarnationNumber > currentListMember.IncarnationNumber {
			*currentListMember = receivedMember
			currentListMember.TimeLocal = now
			// log this update
			common.Logger.Printf("Updated member due to higher incarnation number: %+v\n", currentListMember)
			continue
		}

		if receivedMember.IncarnationNumber == currentListMember.IncarnationNumber {
			if currentListMember.MachineId == self {
				if receivedMember.SuspicionState == common.StateSuspicious && currentListMember.SuspicionState == common.StateAlive {
					// some other machine suspects us, but we know we are alive
					currentListMember.IncarnationNumber++
					currentListMember.TimeLocal = now
					currentListMember.SuspicionState = common.StateAlive
					common.Logger.Printf("Received gossip marking self as suspicious, incremented incarnation number: %+v\n", currentListMember)
					continue
				}
			} else {
				// this means both are alive
				if receivedMember.SuspicionState == common.StateAlive && currentListMember.SuspicionState == common.StateAlive {
					// take the higher heartbeat
					// dont know what to do abt inc # just yet, figure that out later
					common.Logger.Println("Both are alive!")
					if receivedMember.HeartbeatCounter > currentListMember.HeartbeatCounter {
						currentListMember.HeartbeatCounter = receivedMember.HeartbeatCounter
						currentListMember.TimeLocal = now
						common.Logger.Printf("Updated member due to higher heartbeat: %+v\n", currentListMember)
					}
					continue
				}

			}
		}
	}

	common.Logger.Printf("Membership list after merging gossip:")
	for _, member := range list.GetEntireList() {
		common.Logger.Printf("  %s\n", member)
	}
}
