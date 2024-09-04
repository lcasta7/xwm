package utils

import (
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type X11Interface struct {
	conn *xgb.Conn
	root xproto.Window
}

func CreateNewX11Interface(AppCodes map[xproto.Keycode]string) *X11Interface {
	conn, err := xgb.NewConn()
	if err != nil {
		log.Fatalf("Failed to connect to X server: %v", err)
	}

	setup := xproto.Setup(conn)
	if setup == nil {
		log.Fatalf("Failed to get setup information from X server")

	}

	root := setup.DefaultScreen(conn).Root
	if root == 0 {
		log.Fatalf("Failed to get the root window from the default screen")
	}

	xutil := &X11Interface{conn: conn, root: root}
	xutil.initialize(AppCodes)
	return xutil
}

func (x *X11Interface) WaitForEvent() xgb.Event {
	ev, err := x.conn.WaitForEvent()

	if err != nil {
		log.Fatalf("Error waiting for event: %v", err)
	}

	return ev
}

func (x *X11Interface) CleanUp(apps *AppCollection) {
	allWinIds := x.getAllWindowIds()

	appsToRemove := []xproto.Window{}
	for _, v := range apps.AllApps {
		if !slices.Contains(allWinIds, v.WindId) {
			appsToRemove = append(appsToRemove, v.WindId)
		}
	}

	for _, v := range appsToRemove {
		apps.RemoveAppFromCollection(v)
	}
}

// get the WM_Class of the windId at Create time
func (x *X11Interface) GetWMClass(winId xproto.Window) (string, error) {
	wmClassAtom, err := xproto.InternAtom(x.conn, true, uint16(len("WM_CLASS")), "WM_CLASS").Reply()
	if err != nil {
		return "", fmt.Errorf("failed to get WM_CLASS atom: %v", err)
	}

	prop, err := xproto.GetProperty(x.conn, false, winId, wmClassAtom.Atom, xproto.AtomString, 0, (1<<32)-1).Reply()
	if err != nil {
		return "", fmt.Errorf("failed to get WM_CLASS property: %v", err)
	}

	if prop.ValueLen == 0 {
		return "", fmt.Errorf("WM_CLASS not set for window %d", winId)
	}

	classBytes := prop.Value

	nullIndex := strings.IndexByte(string(classBytes), 0)
	if nullIndex == -1 {
		return "", fmt.Errorf("invalid WM_CLASS format")
	}

	className := string(classBytes[nullIndex+1:])

	secondNullIndex := strings.IndexByte(className, 0)
	if secondNullIndex != -1 {
		className = className[:secondNullIndex]
	}

	return strings.ToLower(className), nil
}

func (x *X11Interface) GetCurrentlyFocused() xproto.Window {
	focusReply, err := xproto.GetInputFocus(x.conn).Reply()
	if err != nil {
		log.Printf("Could not get focused window err=%s", err)
	}

	return focusReply.Focus

}

func (x *X11Interface) ChangeWindow(window xproto.Window) bool {
	focused := focusWindow(x.conn, window)
	centered := centerMouseOnWindow(x.conn, window)
	raised := raiseWindow(x.conn, window)

	return focused && centered && raised
}

func focusWindow(conn *xgb.Conn, window xproto.Window) bool {
	// Set the input focus to the specified window
	err := xproto.SetInputFocusChecked(
		conn,
		xproto.InputFocusParent, // Revert focus to PointerRoot when the window is destroyed
		window,
		xproto.TimeCurrentTime, // Use the current time for the focus change
	).Check()

	if err != nil {
		return false
	}

	return true
}

func centerMouseOnWindow(conn *xgb.Conn, window xproto.Window) bool {
	// Get the geometry of the window (position, size, etc.)
	geo, err := xproto.GetGeometry(conn, xproto.Drawable(window)).Reply()
	if err != nil {
		return false
	}

	// Calculate the center of the window
	centerX := int16(geo.X + int16(geo.Width)/2)
	centerY := int16(geo.Y + int16(geo.Height)/2)

	// Move the mouse pointer to the center of the window
	err = xproto.WarpPointerChecked(
		conn,
		xproto.WindowNone,
		window,
		0, 0,
		0, 0,
		centerX, centerY,
	).Check()

	if err != nil {
		return false
	}

	return true
}

func raiseWindow(conn *xgb.Conn, window xproto.Window) bool {
	// Configure the window to be raised (StackModeAbove puts it above all other windows)
	err := xproto.ConfigureWindowChecked(
		conn,
		window,
		xproto.ConfigWindowStackMode, // We are configuring the stacking mode
		[]uint32{xproto.StackModeAbove},
	).Check()

	if err != nil {
		return false
	}

	return true
}

func (x *X11Interface) initialize(AppCodes map[xproto.Keycode]string) {
	x.getAllKeys(AppCodes)
	x.subscribeToWindowAttributes()
}

// will get all the keys defined in app_codes
func (x *X11Interface) getAllKeys(AppCodes map[xproto.Keycode]string) {

	for k := range AppCodes {
		err := xproto.GrabKeyChecked(
			x.conn,
			true,
			x.root,
			xproto.ModMaskAny,
			xproto.Keycode(k),
			xproto.GrabModeAsync,
			xproto.GrabModeAsync,
		).Check()

		if err != nil {
			log.Fatalf("Failed to grab key: %v", err)
		}
	}
}

func (x *X11Interface) subscribeToWindowAttributes() {
	err := xproto.ChangeWindowAttributesChecked(
		x.conn,
		x.root,
		xproto.CwEventMask,
		[]uint32{xproto.EventMaskSubstructureNotify},
	).Check()

	if err != nil {
		log.Fatalf("Failed to subscribe to window creation events: %v", err)
	}
}

func (x *X11Interface) getAllWindowIds() []xproto.Window {
	output, err := exec.Command("wmctrl", "-lx").Output()
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(output), "\n")
	openedWindows := []xproto.Window{}
	for _, line := range lines {
		if line != "" {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				windowId, err := strconv.ParseUint(fields[0], 0, 32)
				if err != nil {
					log.Printf("Could not convert window to xproto.Window window=%v", fields[0])
				}

				openedWindows = append(openedWindows, xproto.Window(windowId))
			}
		}
	}
	return openedWindows
}
