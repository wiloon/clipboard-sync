package main

import (
	"github.com/atotto/clipboard"
	"fmt"
	"strings"
	"sync"
	"time"
	"github.com/samuel/go-zookeeper/zk"
)

func init() {

}

var serverHost string = "172.16.4.210"

func main() {
	//start goroutine to send the content of local clipboard to mq
	var wg sync.WaitGroup
	wg.Add(2)
	go remoteClipboardChangeMonitoe(&wg)

	go localClipboardChangeMonitor(&wg)

	wg.Wait()
	//check local clipboard
	//send to mq

	// start goroutine to receive the message from mq
	// check if update local clipboard
	//update local clipboard
}

// remote
func remoteClipboardChangeMonitoe(wg *sync.WaitGroup) {
	defer wg.Done()
	conn, _, err := zk.Connect([]string{serverHost}, time.Second) //*10)
	if err != nil {
		panic(err)
	}
	lastRemoteClipboard := ""
	for {
		arr, state, err := conn.Get("/clipboard")
		if err != nil {
			fmt.Println("state:", state)
		}
		remoteClipboard := string(arr)
		localClipboard, err := clipboard.ReadAll()
		fmt.Printf("remote clipboard:%v, local clipboard:%v\n", remoteClipboard, localClipboard)
		if !strings.EqualFold(remoteClipboard, lastRemoteClipboard) {
			lastRemoteClipboard = remoteClipboard
			if !strings.EqualFold(remoteClipboard, localClipboard) {
				clipboard.WriteAll(remoteClipboard)
				fmt.Println("remote clipboard change,content:", remoteClipboard)
			}

		}
		time.Sleep(1000 * time.Millisecond)
	}

}

// local
func localClipboardChangeMonitor(wg *sync.WaitGroup) {
	defer wg.Done()

	var tmpMsg string
	conn, _, err := zk.Connect([]string{serverHost}, time.Second) //*10)
	if err != nil {
		panic(err)
	}
	for {
		localClipboard, err := clipboard.ReadAll()
		fmt.Println("local clipboard: ", localClipboard)
		if err != nil {
			fmt.Println("failed to read clipboard:", err)
		}
		if localClipboard != ""&& !strings.EqualFold(localClipboard, tmpMsg) {
			tmpMsg = localClipboard
			fmt.Println("local clipboard change, content:", localClipboard)

			_, state, e := conn.Get("/clipboard")
			if e != nil {

			}

			conn.Set("/clipboard", []byte(localClipboard), state.Version)
		}

		time.Sleep(1000 * time.Millisecond)
	}
}

