package gossip

import (
	"cs425_g12/common"
	"encoding/json"
	"fmt"
	"time"
)

// marshals the data into json
func helperMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Println("Error marshaling:", err)
		return nil
	}
	return data
}

// determines which protocol to run based on the protocol mode
func StartProtocol(list *common.MembershipList, Tgossip time.Duration, Tping time.Duration, Tfail time.Duration) {
	go func() {
		for {
			if common.IsMemberInGroup && common.GetProtocolMode() {
				common.Logger.Println("Running in pingack mode")
				StartPinging(list, Tfail)
				// sleeping to avoid infite loop
				time.Sleep(Tping)
			} else if common.IsMemberInGroup && !common.GetProtocolMode() {
				common.Logger.Println("Running in gossip mode")
				SendGossip(list, Tgossip)
				// sleeping to avoid infite loop
				time.Sleep(Tgossip)
			} else {
				// sleeping to avoid infite loop
				time.Sleep(Tgossip)
				continue
			}
		}
	}()
}
