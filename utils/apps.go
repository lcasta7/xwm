package utils

import (
	"github.com/BurntSushi/xgb/xproto"
	"log"
)

type App struct {
	WindId xproto.Window
	next   *App
	prev   *App
}

type AppCollection struct {
	Size    int
	Current *App
	AllApps []*App
}

func CreateNewAppCollection(win xproto.Window) *AppCollection {
	// initialize the first app
	firstApp := &App{WindId: win}
	firstApp.next = firstApp
	firstApp.prev = firstApp

	// circular list
	collection := &AppCollection{Current: firstApp, Size: 1}
	collection.AllApps = append(collection.AllApps, firstApp)
	return collection
}

// this method will add a node before what's in current for the collection
func (l *AppCollection) AddBeforeCurrent(newId xproto.Window) {
	newApp := &App{WindId: newId}

	defer func() {
		if r := recover(); r != nil {
			log.Print("Recovered from panic: ", r)
			log.Printf("newId=%X l.Current=%X", newId, l.Current.WindId)
			return
		}
	}()

	newApp.prev = l.Current.prev
	newApp.next = l.Current

	l.Current.prev.next = newApp
	l.Current.prev = newApp

	l.Size++
	l.Current = newApp
	l.AllApps = append(l.AllApps, newApp)
}

// this method will add a node before what's in current for the collection
func (l *AppCollection) GoToNextApp(focused xproto.Window) xproto.Window {
	l.findApp(focused)
	l.Current = l.Current.next
	return l.Current.WindId
}

func (l *AppCollection) RemoveAppFromCollection(toRemoveWinId xproto.Window) {
	// Handle case where there's only one element in the list
	if len(l.AllApps) == 1 && l.Current.WindId == toRemoveWinId {
		l.Current = nil
		l.AllApps = nil
		l.Size = 0
		return
	}

	for i, v := range l.AllApps {
		if v.WindId == toRemoveWinId {
			// Update the linked list pointers
			v.prev.next = v.next
			v.next.prev = v.prev

			// Update Current if necessary
			if l.Current == v {
				l.Current = v.next
			}

			// Remove the element from the slice
			l.AllApps = append(l.AllApps[:i], l.AllApps[i+1:]...)

			// Update list size
			l.Size--

			return
		}
	}
}

func (l *AppCollection) ContainsApp(winId xproto.Window) bool {
	for _, a := range l.AllApps {
		if a.WindId == winId {
			return true
		}
	}
	return false
}

func (l *AppCollection) findApp(winId xproto.Window) {
	for l.Current.WindId != winId {
		l.Current = l.Current.next
	}
}
