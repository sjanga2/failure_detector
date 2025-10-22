package hydfs_utils

import (
	"fmt"
	"net/rpc"
)

// sends a local file to another hydfs node over tco
func SendFileToNode(targetAddr string, fileId string, localPath string) error {
	// connect to target rpc server
	client, err := rpc.Dial("tcp", targetAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to HyDFS RPC server at %s: %v", targetAddr, err)
	}
	defer client.Close()

	// read local file into memory
	data, err := ReadLocalFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file %s: %v", localPath, err)
	}

	// make rpc args to send
	args := &FileTransferArgs{
		FileId:   fileId,
		FileData: data,
	}

	// make rpc call
	var reply string
	if err := client.Call("HyDFSReceiver.ReceiveFileFromNode", args, &reply); err != nil {
		return fmt.Errorf("failed to send file %s to %s: %v", fileId, targetAddr, err)
	}

	HyDFSLogger.Printf("Successfully sent file %s to %s: %s", fileId, targetAddr, reply)
	return nil
}