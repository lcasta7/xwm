package main

import (
	"fmt"
	"log"
	"os/exec"
	"sync"

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
		xproto.Keycode(75),   // Keycode for F12 (change if necessary)
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
							if v.collection != nil {
								fmt.Printf("Key: %v, Content: 0x%v\n", k, v.collection.ToSlice())
							}
						}
					}
				//case 75:
				//	fallthrough
				case 69:
					{

						if value, exist := apps[app_codes[event.Detail]]; exist && len(value.collection.ToSlice()) != 0 {

							if focusedWindow := GetCurrentlyFocused(conn); focusedWindow != 0 {
								if ContainsApp(value.collection.ToSlice(), focusedWindow) &&
									value.current.Next != nil && value.current.Next.Data != 0 {
									value.current = value.current.Next
								}

								ChangeWindow(conn, value.current.Data)
							}
						} else {
							log.Printf("Creating new Client=%v\n", event.Detail)
							if created := CreateClient(event.Detail, conn, root); len(created) != 0 {
								last_launched = created
								apps[last_launched] = nil
							}
						}

					}
				}
			}

		case xproto.DestroyNotifyEvent: //here
			{

				apps = ClearApps(GetActualWindowIds(conn, root), apps, event.Window)
				TidyUp(GetActualWindowIds(conn, root), apps)
			}

		case xproto.CreateNotifyEvent:
			{
				log.Printf("adding app=%s\n", last_launched)
				if _, exist := apps[last_launched]; exist {
					AddToAppsList(last_launched, event)
				}
			}
		}
	} //end of infinite loop

}

func CreateClient(keyCode xproto.Keycode, conn *xgb.Conn, root xproto.Window) string {

	if value, exist := app_codes[keyCode]; exist {
		cmd := exec.Command(value)
		err := cmd.Start()
		log.Print("Waiting for app\n")
		if err != nil {
			log.Printf("Failed to launch program: %v", err)
		} else {
			//var wg sync.WaitGroup
			//tidyUp := func(wg *sync.WaitGroup) {
			//	defer wg.Done()
			//	mu.Lock()
			//	time.Sleep(3 * time.Second)
			//	TidyUp(GetActualWindowIds(conn, root), apps)
			//	mu.Unlock()
			//}
			//wg.Add(1)
			//go tidyUp(&wg)
			//TidyUp(GetActualWindowIds(conn, root), apps)
			return value
		}
	} else {
		log.Println("program not in config.go, please add it first")
	}

	return ""
}

func ChangeWindow(conn *xgb.Conn, window xproto.Window) bool {
	log.Printf("new window=%X and %v\n", window, window)
	focused, _ := FocusWindow(conn, window)
	centered, _ := CenterMouseOnWindow(conn, window)
	raised, _ := RaiseWindow(conn, window)

	log.Printf("values for focused=%v centered=%v, raised=%v", focused, centered, raised)
	return focused && centered && raised
}

func RaiseWindow(conn *xgb.Conn, window xproto.Window) (bool, error) {
	// Configure the window to be raised (StackModeAbove puts it above all other windows)
	err := xproto.ConfigureWindowChecked(
		conn,
		window,
		xproto.ConfigWindowStackMode, // We are configuring the stacking mode
		[]uint32{xproto.StackModeAbove},
	).Check()

	if err != nil {
		return false, err
	}

	return true, nil
}

func CenterMouseOnWindow(conn *xgb.Conn, window xproto.Window) (bool, error) {
	// Get the geometry of the window (position, size, etc.)
	geo, err := xproto.GetGeometry(conn, xproto.Drawable(window)).Reply()
	if err != nil {
		return false, fmt.Errorf("failed to get window geometry: %w", err)
	}

	// Calculate the center of the window
	centerX := int16(geo.X + int16(geo.Width)/2)
	centerY := int16(geo.Y + int16(geo.Height)/2)

	// Move the mouse pointer to the center of the window
	err = xproto.WarpPointerChecked(
		conn,
		xproto.WindowNone, // The source window (None means no constraints)
		window,            // The destination window
		0, 0,              // Source coordinates (not relevant when WindowNone is used)
		0, 0, // Source width and height (not relevant when WindowNone is used)
		centerX, centerY, // Destination coordinates (center of the window)
	).Check()

	if err != nil {
		return false, fmt.Errorf("failed to warp pointer: %w", err)
	}

	return true, nil
}

func FocusWindow(conn *xgb.Conn, window xproto.Window) (bool, error) {
	// Set the input focus to the specified window
	err := xproto.SetInputFocusChecked(
		conn,
		xproto.InputFocusPointerRoot, // Revert focus to PointerRoot when the window is destroyed
		window,
		xproto.TimeCurrentTime, // Use the current time for the focus change
	).Check()

	if err != nil {
		return false, err
	}

	return true, nil
}

func ContainsApp(s []xproto.Window, e xproto.Window) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GetCurrentlyFocused(conn *xgb.Conn) xproto.Window {
	conn, err := xgb.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	focusReply, err := xproto.GetInputFocus(conn).Reply()
	if err != nil {
		log.Fatal(err)
	}

	return focusReply.Focus

}

func AddToAppsList(last_launched string, event xproto.CreateNotifyEvent) {
	app := &app{win: last_launched, win_id: event.Window}

	if app_col, exists := apps[app.win]; exists && app_col != nil {

		new_node, new_collection := app_col.collection.AddBetween(app.win_id,
			app_col.current.Prev,
			app_col.current)

		app_col.current = new_node
		app_col.collection = new_collection

		apps[app.win] = app_col
	} else {
		new_collection := NewList[xproto.Window]()
		new_group := &app_collection{
			collection: &new_collection,
		}

		log.Printf("Creating New Group\n")
		new_node := new_group.collection.AddFirst(app.win_id)
		new_group.current = new_node
		apps[app.win] = new_group
	}
}
