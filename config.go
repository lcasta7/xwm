package main

import (
	"log"
	"slices"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

var app_codes = map[xproto.Keycode]string{
	69: "vivaldi-stable",
	65: "emacs",
}

func ClearApps(actual_window_ids []xproto.Window, apps map[string]*app_collection, deletedWin xproto.Window) map[string]*app_collection {

	log.Printf("Starting ClearApps Job")
	if len(apps) == 0 {
		log.Printf("no apps to clean\n")
		return apps
	}

	//make a copy of the apps
	temp_apps := make(map[string]*app_collection)
	for k, v := range apps {
		temp_apps[k] = v
	}

	keys_to_delete := []xproto.Window{}
	for _, v := range apps {
		if v.collection != nil {
			continue
		}
		app_list := v.collection.ToSlice()
		for _, app := range app_list {
			if !slices.Contains(actual_window_ids, app) {
				keys_to_delete = append(keys_to_delete, app)
			}
		}
	}

	//need to change current
	keys_to_delete = append(keys_to_delete, deletedWin)

	//delete the keys
	for k, v := range temp_apps {
		for _, to_delete_app := range keys_to_delete {
			temp_list, temp_node := v.collection.RemoveFirstFound(to_delete_app)
			if temp_node == nil {
				delete(temp_apps, k)
			} else {
				temp_apps[k].collection = temp_list
				temp_apps[k].current = temp_node
			}
		}
	}

	return temp_apps
}

func GetActualWindowIds(conn *xgb.Conn, root xproto.Window) []xproto.Window {
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

	return actual_window_ids
}

func TidyUp(actual_window_ids []xproto.Window, apps map[string]*app_collection) []xproto.Window {

	log.Printf("Starting TidyUp Job")
	for _, v := range apps {
		// this can probably be simplified
		for v.current != nil &&
			v.current.Next != nil &&
			!v.collection.IsLast(v.current) {
			if !slices.Contains(actual_window_ids, v.current.Data) {
				old_node := v.current
				v.current = v.current.Next
				v.current.Prev = old_node.Prev
				v.collection.Remove(old_node)
			} else {
				v.current = v.current.Next
			}
		}
	}

	return actual_window_ids
}
