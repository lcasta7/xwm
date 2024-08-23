package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
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

func TidyUp(conn *xgb.Conn, root xproto.Window, wins map[string]*app_collection) {
	treeReply, err := xproto.QueryTree(conn, root).Reply()
	if err != nil {
		log.Fatal(err)
	}

	//collect all the windows
	actual_window_ids := []xproto.Window{}
	wmStateAtom, err := xproto.InternAtom(conn, true, 8, "WM_STATE").Reply()
	for _, window := range treeReply.Children {

		// Skip windows without WM_STATE property
		propReply, err := xproto.GetProperty(conn, false, window, wmStateAtom.Atom, xproto.GetPropertyTypeAny, 0, 1).Reply()
		if err != nil || len(propReply.Value) == 0 {
			continue
		}

		actual_window_ids = append(actual_window_ids, window)
	}

	fmt.Print("starting cleanup\n")
	for _, v := range wins {
		for v.current != nil && v.current.Next != nil && !v.collection.IsLast(v.current) {
			fmt.Printf("actual_wind_ids=%v, v.current.Data=0x%X\n", actual_window_ids, v.current.Data)
			if !slices.Contains(actual_window_ids, v.current.Data) {
				old_node := v.current
				v.current = v.current.Next
				v.collection.Remove(old_node)
			} else {
				v.current = v.current.Next
			}
		}
	}

	fmt.Print("Done cleaning\n")
}

func restart() *exec.Cmd {
	log.Println("Restarting")

	// Define the command and arguments separately
	exePath := "go"
	args := []string{"run", "/home/luis/dev/xwm/xwm.go", "/home/luis/dev/xwm/list.go"}

	log.Printf("The command: %s %v", exePath, args)

	// Prepare the command with the specified arguments
	cmd := exec.Command(exePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Exit the current instance of the program
	return cmd
}

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
	apps := make(map[string]*app_collection)
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

						for k, v := range apps {
							fmt.Printf("Key: %v, Content: 0x%X\n", k, v.collection.ToSlice())

						}

					}
				case 82:
					{
						defer restart().Run()
						apps = make(map[string]*app_collection)
						break
					}
				default:
					{
						if value, exist := app_codes[event.Detail]; exist {
							cmd := exec.Command(value)
							err := cmd.Start()
							log.Print("Waiting for app\n")
							if err != nil {
								log.Printf("Failed to launch program: %v", err)
							} else {
								log.Printf("Launched program successfully %v\n", value)
								last_launched = value
								go func() {
									time.Sleep(2 * time.Second)
									TidyUp(conn, root, apps)
								}()
							}

						} else {
							log.Println("program not in config.go, please add it first")
						}
					}
				}
			}
		case xproto.CreateNotifyEvent:
			{
				app := &app{win: last_launched, win_id: event.Window}

				fmt.Print("----------start------------\n")
				if app_col, exists := apps[app.win]; exists {
					fmt.Print("adding existing\n")
					new_node := app_col.collection.AddBetween(app.win_id,
						app_col.current.Prev,
						app_col.current)

					app_col.current = new_node
					fmt.Printf("Adding to exciting group=0x%X", app.win)
				} else {

					fmt.Print("adding new\n")
					new_collection := NewList[xproto.Window]()
					new_group := &app_collection{
						collection: &new_collection,
					}

					new_node := new_group.collection.AddFirst(app.win_id)
					new_group.current = new_node
					apps[app.win] = new_group

					fmt.Printf("Created new group=%v\n", app.win)
				}

				fmt.Printf("still working?\n")
			}

		}
	}
}
