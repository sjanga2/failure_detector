package gossip

import (
	"cs425_g12/common"
	"fmt"
	"time"
)

func MergeGossip(list *common.MembershipList, receivedGossip []common.Member, self common.MachineId) {
	common.Logger.Printf("Merge function entered! Received gossip: %+v\n", receivedGossip)

	now := time.Now()

	for _, receivedMember := range receivedGossip {
		currentListMember := list.GetMember(receivedMember.MachineId)

		if currentListMember == nil {
			// new member adding to the list, only if it is not a failed member (to prevent ghost entries)
			if receivedMember.SuspicionState != common.StateFailed {
				receivedMember.TimeLocal = now
				list.Insert(receivedMember)
				common.Logger.Printf("Added new member: %+v\n", receivedMember)
			}
			continue
		}

		if receivedMember.SuspicionState == common.StateFailed {
			// failure overrides everything
			if currentListMember.SuspicionState != common.StateFailed {
				fmt.Printf("[%s] Member %+v marked as Failed (from gossip)\n", time.Now().Format("15:04:05.000"), currentListMember.MachineId)
				common.Logger.Printf("DropSearch First notification of failure for member: %+v\n", currentListMember)
			}
			currentListMember.HeartbeatCounter = receivedMember.HeartbeatCounter
			currentListMember.SuspicionState = common.StateFailed
			currentListMember.IncarnationNumber = receivedMember.IncarnationNumber
			// log this update
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

		// same inc number, sus takes priority over alive
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
				// always check heartbeat first if same incarnation number
				if receivedMember.HeartbeatCounter > currentListMember.HeartbeatCounter {
					// received definitely has more recent info, update it
					currentListMember.HeartbeatCounter = receivedMember.HeartbeatCounter
					currentListMember.TimeLocal = now
					currentListMember.SuspicionState = receivedMember.SuspicionState
					common.Logger.Printf("Updated member due to higher heartbeat: %+v\n", currentListMember)
				} else if receivedMember.HeartbeatCounter == currentListMember.HeartbeatCounter {
					// same heartbeat, so sus takes priority over alive
					if receivedMember.SuspicionState == common.StateSuspicious && currentListMember.SuspicionState == common.StateAlive {
						currentListMember.SuspicionState = common.StateSuspicious
						currentListMember.TimeLocal = now
						common.Logger.Printf("Updated member to Suspicious due to same heartbeat but received is suspicious: %+v\n", currentListMember)
					} // else do nothing, either both are sus or current is sus
				} // else received hearbeat < current heartbeat, ignore that
			}
		}
	}

	common.Logger.Printf("Membership list after merging gossip:")
	for _, member := range list.GetEntireList() {
		common.Logger.Printf("  %s\n", member)
	}
}
