package utils

import (
	"testing"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/stretchr/testify/assert"
)

func TestCreateNewAppCollection(t *testing.T) {
	win := xproto.Window(123)
	appCollection := CreateNewAppCollection(win)

	assert.Equal(t, 1, appCollection.Size)
	assert.Equal(t, appCollection.Current, appCollection.Current.next)
	assert.Equal(t, appCollection.Current, appCollection.Current.prev)
	assert.Equal(t, 1, len(appCollection.AllApps))
}

func TestAddBeforeCurrent(t *testing.T) {
	win := xproto.Window(123)
	appCollection := CreateNewAppCollection(win)

	appCollection.AddBeforeCurrent(456)

	assert.Equal(t, 2, appCollection.Size)
	assert.Equal(t, xproto.Window(456), appCollection.Current.WindId)
	assert.Equal(t, xproto.Window(123), appCollection.Current.next.WindId)
	assert.Equal(t, xproto.Window(456), appCollection.Current.next.next.WindId)
	assert.Equal(t, xproto.Window(123), appCollection.Current.next.next.next.WindId)
	assert.Equal(t, xproto.Window(456), appCollection.Current.next.next.next.next.WindId)
	assert.Equal(t, 2, len(appCollection.AllApps))
}

func TestRemoveAppFromCollection(t *testing.T) {
	//var empty_app *App = nil
	win := xproto.Window(123)
	appCollection := CreateNewAppCollection(win)

	appCollection.AddBeforeCurrent(456)
	appCollection.AddBeforeCurrent(457)
	appCollection.AddBeforeCurrent(458)

	assert.Equal(t, 4, appCollection.Size)
	assert.Equal(t, xproto.Window(458), appCollection.Current.WindId)
	assert.Equal(t, xproto.Window(457), appCollection.Current.next.WindId)
	assert.Equal(t, xproto.Window(456), appCollection.Current.next.next.WindId)

	appCollection.RemoveAppFromCollection(458)
	assert.Equal(t, xproto.Window(457), appCollection.Current.WindId)

	appCollection.RemoveAppFromCollection(457)
	assert.Equal(t, xproto.Window(456), appCollection.Current.WindId)

	appCollection.RemoveAppFromCollection(456)
	assert.Equal(t, xproto.Window(123), appCollection.Current.WindId)

}

func TestContainsApp(t *testing.T) {
	win := xproto.Window(123)
	appCollection := CreateNewAppCollection(win)

	assert.False(t, appCollection.ContainsApp(xproto.Window(456)))
	appCollection.AddBeforeCurrent(456)

	assert.True(t, appCollection.ContainsApp(xproto.Window(123)))
	assert.True(t, appCollection.ContainsApp(xproto.Window(456)))
}
