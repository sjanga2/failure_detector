package common

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// constant port for dialing in the machines (introducer is also on this port)
const GlobalPort = 5051

var Logger *log.Logger

// logger file
func InitializeLogger(machineName string) {
	logPath := filepath.Join("/home/shared", machineName+".log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("could not open log file: %v", err)
	}
	Logger = log.New(file, "", log.Ldate|log.Ltime|log.Lmicroseconds)
}

// used for starting the protocols only when the members are part of the group
var IsMemberInGroup bool

// self struct - global
var (
	Self     MachineId
	SelfLock sync.RWMutex
)

func SetSelf(m MachineId) {
	SelfLock.Lock()
	defer SelfLock.Unlock()
	Self = m
}

func GetSelf() MachineId {
	SelfLock.RLock()
	defer SelfLock.RUnlock()
	return Self
}

// part of membership list entry specifying the machine
type MachineId struct {
	Ip      string
	Port    uint16
	Version int64 // version of the machine
}

// constructor for machineId
func NewMachineId(ip string, port uint16, startTime time.Time) MachineId {
	return MachineId{Ip: ip, Port: port, Version: startTime.UnixNano()}
}

// introducer machine : vm 1
var Introducer MachineId

// state tracker for machine
type SuspicionState uint8

// state machine for the suspicion states
const (
	StateAlive SuspicionState = iota
	StateSuspicious
	StateFailed
)

// member entries struct to add to membership list
type Member struct {
	MachineId         MachineId
	HeartbeatCounter  uint64
	TimeLocal         time.Time // last updated time
	SuspicionState    SuspicionState
	IncarnationNumber uint64 // incarnation number

	RingId       [20]byte
	RingIdString string
}

// constructor for member
func NewMember(machineId MachineId) Member {

	idStr := fmt.Sprintf("%s:%d:%d", machineId.Ip, machineId.Port, machineId.Version)
	hash := sha1.Sum([]byte(idStr))
	// directly using the hash here, assuming m = 160
	ringIDStr := hex.EncodeToString(hash[:])

	return Member{
		MachineId:         machineId,
		HeartbeatCounter:  0,
		TimeLocal:         time.Now(),
		SuspicionState:    StateAlive,
		IncarnationNumber: 0,
		RingId:            hash,
		RingIdString:      ringIDStr,
	}
}

// membership list struct
type MembershipList struct {
	members    map[MachineId]*Member
	sortedRing []*Member
	mutex      sync.RWMutex
}

// constructor for membership list
func NewMembershipList() *MembershipList {
	return &MembershipList{
		members:    make(map[MachineId]*Member),
		sortedRing: make([]*Member, 0),
	}
}

// pretty printing functions
func (m MachineId) String() string {
	return fmt.Sprintf("%s:%d (v%d)", m.Ip, m.Port, m.Version)
}

func (m Member) String() string {
	return fmt.Sprintf(
		"[ID=%s | HB=%d | Inc=%d | State=%s | LastSeen=%s | RingId=%d | RingIdString=%s]",
		m.MachineId,
		m.HeartbeatCounter,
		m.IncarnationNumber,
		m.SuspicionState.String(),
		m.TimeLocal.Format("15:04:05.000"),
		m.RingId,
		m.RingIdString,
	)
}

func (s SuspicionState) String() string {
	switch s {
	case StateAlive:
		return "Alive"
	case StateSuspicious:
		return "Suspicious"
	case StateFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// sorting the sortedRing based on RingId
func (list *MembershipList) updateSortedRing() {
	list.sortedRing = list.sortedRing[:0]
	for _, m := range list.members {
		list.sortedRing = append(list.sortedRing, m)
	}
	sort.Slice(list.sortedRing, func(i, j int) bool {
		return bytes.Compare(list.sortedRing[i].RingId[:], list.sortedRing[j].RingId[:]) < 0
	})
}

// insert member into membership list
func (list *MembershipList) Insert(member Member) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	// check if already existing?
	_, exists := list.members[member.MachineId]
	if !exists {
		//if it does not exist, then we add it into the list
		list.members[member.MachineId] = &member
		list.updateSortedRing()
	}

	Logger.Printf("Inserted member: %+v\n", member)
	Logger.Println("Membership list after insertion:")
	for _, m := range list.members {
		Logger.Printf("   %s\n", *m)
	}
}

// delete one member from the list
func (list *MembershipList) Delete(MachineId MachineId) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	delete(list.members, MachineId)
	list.updateSortedRing()
	Logger.Printf("Deleted member: %+v\n", MachineId)
}

// delete full membership list used for cleanup
func (list *MembershipList) DeleteEntireList() {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	for k := range list.members {
		delete(list.members, k)
	}
	list.sortedRing = list.sortedRing[:0]
	Logger.Println("Cleared entire membership list")
}

// returns the entire list
func (list *MembershipList) GetEntireList() []Member {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	out := make([]Member, 0, len(list.members))
	for _, member := range list.members {
		out = append(out, *member)
	}
	return out
}

// for accessing the unique members in the list (used when choosing target)
func (list *MembershipList) GetUniqueMembers() []Member {
	members := list.GetEntireList()
	seen := make(map[string]bool) // mark each ip as seen or not seen
	out := make([]Member, 0, len(members))

	Logger.Printf("Creating the alive list...")
	for _, member := range members {
		// if member.SuspicionState == StateFailed {
		// 	continue
		// }
		key := fmt.Sprintf("%s:%d", member.MachineId.Ip, member.MachineId.Port)
		if !seen[key] {
			seen[key] = true
			out = append(out, member)
			Logger.Printf("Alive member: %+v", member)
		}
	}
	return out
}

// returns the member if it exists
func (list *MembershipList) GetMember(machine MachineId) *Member {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	if member, exists := list.members[machine]; exists {
		// memberCopy := *member
		// return &memberCopy
		return member
	}
	return nil
}

// return the sorted ring array
func (list *MembershipList) GetSortedRing() []*Member {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return append([]*Member(nil), list.sortedRing...)
}

// FindSuccessor function
func (list *MembershipList) FindSuccessor(ringID [20]byte) *Member {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	if len(list.sortedRing) == 0 {
		return nil // no members in ring
	}

	for _, member := range list.sortedRing {
		if bytes.Compare(member.RingId[:], ringID[:]) > 0 { // only want > not >= since we dont want exact match (node itself)
			return member
		}
	}
	return list.sortedRing[0] // if wrap around
}

// FindPredecessor function
func (list *MembershipList) FindPredecessor(ringID [20]byte) *Member {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	if len(list.sortedRing) == 0 {
		return nil
	}

	// walk through sorted ring, keep last member smaller than ID
	var pred *Member
	for _, member := range list.sortedRing {
		if bytes.Compare(member.RingId[:], ringID[:]) < 0 {
			pred = member
		} else {
			break
		}
	}

	// wrap-around case
	if pred == nil {
		pred = list.sortedRing[len(list.sortedRing)-1]
	}
	return pred
}

// get successor nodes function
func (list *MembershipList) GetSuccessorNodes(fileID [20]byte, n int) []*Member {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	successors := make([]*Member, 0, n)
	count := len(list.sortedRing)
	if count == 0 {
		return successors
	}

	if n > count {
		n = count
	}

	// need next n successors
	startIndex := -1
	for i, member := range list.sortedRing {
		if bytes.Compare(member.RingId[:], fileID[:]) > 0 {
			startIndex = i
			break
		}
	}
	if startIndex == -1 {
		startIndex = 0 // wrap around
	}

	for i := range n {
		index := (startIndex + i) % count
		successors = append(successors, list.sortedRing[index])
	}
	return successors
}

// heartbeat counter incrementor
func (list *MembershipList) IncrementHeartbeat() {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	found := false
	for _, member := range list.members {
		// only want to increment heartbeat to the members not marked as failed
		if member.MachineId.Ip == GetSelf().Ip {
			if member.SuspicionState != StateFailed {
				member.HeartbeatCounter++
				member.TimeLocal = time.Now()
				// member.SuspicionState = StateAlive
				Logger.Printf("Incremented heartbeat for member: %+v\n", member)
			}
			found = true
			break
		}
	}

	if !found {
		Logger.Printf("Machine with IP %s not found in membership list. Cannot increment heartbeat.\n", GetSelf().Ip)
	}
}

// sus vs no sus
var (
	useSuspicion bool
	modeMutex    sync.RWMutex
)

func SetSuspicionMode(susValue bool) { // if susValue is true, then run suscheck
	modeMutex.Lock()
	defer modeMutex.Unlock()
	useSuspicion = susValue
}

func GetSuspicionMode() bool {
	modeMutex.Lock()
	defer modeMutex.Unlock()
	return useSuspicion
}

// pingack vs gossip
var (
	usePingAck    bool
	protocolMutex sync.RWMutex
)

func SetProtocolMode(mode bool) {
	protocolMutex.Lock()
	defer protocolMutex.Unlock()
	usePingAck = mode
}

func GetProtocolMode() bool {
	protocolMutex.Lock()
	defer protocolMutex.Unlock()
	return usePingAck
}

// starts the checker based on the mode
func (list *MembershipList) StartChecker(Tsus time.Duration, Tfail time.Duration, Tclean time.Duration, Tsuscheck time.Duration, Tfailcheck time.Duration) {
	go func() {
		for {
			if GetSuspicionMode() {
				list.StartSuspicionChecker(Tsus, Tfail, Tclean, Tsuscheck)
				time.Sleep(Tsuscheck)
			} else {
				list.StartFailedChecker(Tfail, Tclean, Tfailcheck)
				time.Sleep(Tfailcheck)
			}
		}
	}()
}

func (list *MembershipList) StartSuspicionChecker(Tsus time.Duration, Tfail time.Duration, Tclean time.Duration, Tsuscheck time.Duration) {
	// go func() {
	// 	for {
	// 		time.Sleep(Tsuscheck)

	now := time.Now()
	list.mutex.Lock()
	for id, member := range list.members {
		if id.Ip == GetSelf().Ip {
			// Logger.Printf("I am self: %+v", GetSelf())
			continue // skip self
		}

		elapsed := now.Sub(member.TimeLocal)

		if !GetProtocolMode() && member.SuspicionState == StateAlive {
			if elapsed > Tsus {
				// member is sus
				member.SuspicionState = StateSuspicious
				fmt.Printf("Member %+v marked as Suspicious, elapsed time: %+v\n", member.MachineId, elapsed)
				Logger.Printf("Member %+v marked as Suspicious, elapsed time: %+v\n", member.MachineId, elapsed)
			} // else still alive, continue being alive
		} else if member.SuspicionState == StateSuspicious {
			if elapsed > Tfail {
				// member has failed
				member.SuspicionState = StateFailed
				fmt.Printf("[%s] Member %+v marked as Failed (timeout in suschecker)\n", time.Now().Format("15:04:05.000"), member.MachineId)
				Logger.Printf("DropSearch Member %+v marked as Failed, elapsed time: %+v\n", member.MachineId, elapsed)
			}
		} else if member.SuspicionState == StateFailed {
			if elapsed > Tclean {
				// remove member from list
				delete(list.members, id)
				Logger.Printf("Member %+v removed from membership list due to cleanup, elapsed time: %+v\n", member.MachineId, elapsed)
			}
		}
	}
	list.mutex.Unlock()
}

// 	}()
// }

func (list *MembershipList) StartFailedChecker(Tfail time.Duration, Tclean time.Duration, Tfailcheck time.Duration) {
	// go func() {
	// 	for {
	// 		time.Sleep(Tfailcheck)
	now := time.Now()
	list.mutex.Lock()

	for id, member := range list.members {
		if id.Ip == GetSelf().Ip {
			// Logger.Printf("I am self: %+v", GetSelf())
			continue // skip self
		}

		elapsed := now.Sub(member.TimeLocal)
		if !GetProtocolMode() && member.SuspicionState == StateAlive {
			if elapsed > Tfail {
				// remove member from list
				member.SuspicionState = StateFailed
				fmt.Printf("[%s] Member %+v marked as Failed (timeout in failchecker)\n", time.Now().Format("15:04:05.000"), member.MachineId)
				Logger.Printf("DropSearch Member %+v marked as Failed, elapsed time: %+v\n", member.MachineId, elapsed)
			}
		} else if member.SuspicionState == StateFailed {
			if elapsed > Tclean {
				delete(list.members, id)
				Logger.Printf("Member %+v removed from membership list due to cleanup, elapsed time: %+v\n", member.MachineId, elapsed)
			}
		}
	}
	list.mutex.Unlock()
}

// 	}()
// }
