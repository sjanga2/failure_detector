package hydfs_utils

import (
	"fmt"
	"net"
	"net/rpc"
)

type FileTransferArgs struct {
	FileName string // user facing filename
	FileId   string // filename is only for user facing functions, every internal function uses fileId
	FileData []byte // is data sent as raw bytes?
}

type HyDFSReceiver struct {
	DataDir string // directory to store received files
}

func (r *HyDFSReceiver) ReceiveFileFromNode(args *FileTransferArgs, reply *string) error {
	destPath := fmt.Sprintf("%s/%s", r.DataDir, args.FileId)
	if err := WriteLocalFile(destPath, args.FileData); err != nil {
		return fmt.Errorf("failed to store file %s: %v", args.FileId, err)
	}
	*reply = fmt.Sprintf("Stored %s (%d bytes)", args.FileId, len(args.FileData)) // any go rpc method has to have a reply
	return nil
}

func InitHyDFS(port string) error {
	if err := InitHyDFSDir(); err != nil {
		return fmt.Errorf("failed to initialize HyDFS directories: %v", err)
	}

	// register the rpc receiver
	receiver := &HyDFSReceiver{DataDir: "/home/shared/hydfs/data"}
	if err := rpc.Register(receiver); err != nil {
		return fmt.Errorf("failed to register HyDFSReceiver RPC: %v", err)
	}

	// start listening for rpc conn
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to start HyDFS RPC listener: %v", err)
	}

	// serve rpc requests in this goroutine
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				HyDFSLogger.Printf("failed to accept HyDFS RPC connection: %v", err)
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()

	return nil
}
