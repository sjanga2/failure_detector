// package main

// import (
// 	"cs425_g12/hydfs_utils"
// 	"fmt"
// )

// func main() {
// 	hydfs_utils.InitHyDFSDir()

// 	data := []byte("Testing HyDFS local file I/O\n")
// 	err := hydfs_utils.WriteLocalFile("/home/shared/hydfs/data/test1.txt", data)
// 	if err != nil {
// 		fmt.Println("Write error:", err)
// 		return
// 	}

// 	readData, err := hydfs_utils.ReadLocalFile("/home/shared/hydfs/data/test1.txt")
// 	if err != nil {
// 		fmt.Println("Read error:", err)
// 		return
// 	}

// 	fmt.Printf("Read back: %s\n", string(readData))
// }

package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"

	"cs425_g12/hydfs_utils"
)

func main() {
	// 1. Create a simple directory for the receiver
	hydfs_utils.InitializeLogger("testingmach1")
	dataDir := "./testdata"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Printf("failed to create dataDir: %v\n", err)
		return
	}
	// defer os.RemoveAll(dataDir)

	// 2. Initialize and register the HyDFSReceiver
	receiver := &hydfs_utils.HyDFSReceiver{DataDir: dataDir}
	if err := rpc.Register(receiver); err != nil {
		fmt.Printf("failed to register receiver: %v\n", err)
		return
	}

	// 3. Listen on a free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
		return
	}
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port
	fmt.Printf("RPC server listening on port %d\n", port)

	// 4. Serve RPC requests in a goroutine
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()

	// 5. Create a tiny file to send
	content := []byte("hi")
	localPath := dataDir + "/localfile.txt"
	if err := os.WriteFile(localPath, content, 0644); err != nil {
		fmt.Printf("failed to write local file: %v\n", err)
		return
	}

	// 6. Send the file
	fileID := "file1"
	targetAddr := "127.0.0.1:" + strconv.Itoa(port)
	if err := hydfs_utils.SendFileToNode(targetAddr, fileID, localPath); err != nil {
		fmt.Printf("SendFileToNode failed: %v\n", err)
		return
	}

	// 7. Verify file was received
	storedPath := dataDir + "/" + fileID
	storedData, err := os.ReadFile(storedPath)
	if err != nil {
		fmt.Printf("failed to read stored file: %v\n", err)
		return
	}

	if string(storedData) != string(content) {
		fmt.Printf("file content mismatch, got %s, want %s\n", string(storedData), string(content))
		return
	}

	fmt.Println("Test succeeded! File sent and received correctly.")
}
