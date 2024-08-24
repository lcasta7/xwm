package main

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type app struct {
	win    string
	win_id xproto.Window
}

type app_collection struct {
	current    *Node[xproto.Window]
	collection *LinkedList[xproto.Window]
}
var (
	mu   sync.Mutex 
	apps map[string]*app_collection 
)
func main() {

	conn, err := xgb.NewConn()
	if err != nil {
		log.Fatalf("Failed to connect to X server: %v", err)
	}
	defer conn.Close()
	// Subscribe to key press events
	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root

	err = xproto.GrabKeyChecked(
		conn,
		true,                 // OwnerEvents
		root,                 // Window
		xproto.ModMaskAny,    // Modifiers
		xproto.Keycode(96),   // Keycode for F12 (change if necessary)
		xproto.GrabModeAsync, // Pointer mode
		xproto.GrabModeAsync, // Keyboard mode
	).Check()

	err = xproto.GrabKeyChecked(
		conn,
		true,                 // OwnerEvents
		root,                 // Window
		xproto.ModMaskAny,    // Modifiers
		xproto.Keycode(82),   // Keycode for F12 (change if necessary)
		xproto.GrabModeAsync, // Pointer mode
		xproto.GrabModeAsync, // Keyboard mode
	).Check()

	err = xproto.GrabKeyChecked(
		conn,
		true,                 // OwnerEvents
		root,                 // Window
		xproto.ModMaskAny,    // Modifiers
		xproto.Keycode(69),   // Keycode for F12 (change if necessary)
		xproto.GrabModeAsync, // Pointer mode
		xproto.GrabModeAsync, // Keyboard mode
	).Check()

	if err != nil {
		log.Fatalf("Failed to grab key: %v", err)
	}

	err = xproto.ChangeWindowAttributesChecked(
		conn,
		root,
		xproto.CwEventMask,
		[]uint32{xproto.EventMaskSubstructureNotify},
	).Check()
	if err != nil {
		log.Fatalf("Failed to subscribe to window creation events: %v", err)
	}

	// map from className to list winId
	apps = make(map[string]*app_collection)
	last_launched := ""
	for {
		ev, err := conn.WaitForEvent()
		if err != nil {
			log.Fatalf("Error waiting for event: %v", err)
		}

		switch event := ev.(type) {
		case xproto.KeyReleaseEvent:
			{
				switch event.Detail {
				case 96:
					{
						if len(apps) == 0 {
							log.Print("no apps to show/n")
						}
						for k, v := range apps {
							fmt.Printf("Key: %v, Content: 0x%v\n", k, v.collection.ToSlice())
						}
					}
				//case 82: {}
				default:
					{
						if value, exist := app_codes[event.Detail]; exist {
							cmd := exec.Command(value)
							err := cmd.Start()
							log.Print("Waiting for app\n")
							if err != nil {
								log.Printf("Failed to launch program: %v", err)
							} else {
								last_launched = value
								var wg sync.WaitGroup
								tidyUp := func(wg *sync.WaitGroup) {
									defer wg.Done()
									mu.Lock()
									time.Sleep(2 * time.Second)
									TidyUp(GetActualWindowIds(conn, root), apps)
									mu.Unlock()
								}
								wg.Add(1)
								go tidyUp(&wg)

							}
						} else {
							log.Println("program not in config.go, please add it first")
						}
					}
				}
			}

		case xproto.DestroyNotifyEvent:
			{

				var wg sync.WaitGroup
				cleanUp := func(wg *sync.WaitGroup) {
					defer wg.Done()
					mu.Lock()
					time.Sleep(1 * time.Second)
					apps = ClearApps(GetActualWindowIds(conn, root), apps)
					mu.Unlock()
				}
				wg.Add(1)
				go cleanUp(&wg)
			}

		case xproto.CreateNotifyEvent:
			{
				app := &app{win: last_launched, win_id: event.Window}

				if app_col, exists := apps[app.win]; exists {
					new_node := app_col.collection.AddBetween(app.win_id,
						app_col.current.Prev,
						app_col.current)

					app_col.current = new_node
				} else {
					new_collection := NewList[xproto.Window]()
					new_group := &app_collection{
						collection: &new_collection,
					}

					new_node := new_group.collection.AddFirst(app.win_id)
					new_group.current = new_node
					apps[app.win] = new_group
				}
			}
		}
	}
}
