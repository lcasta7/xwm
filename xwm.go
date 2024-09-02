package main

import (
	"log"
	"os/exec"
	"slices"
	"time"
	"xwm/revision/utils"

	"xwm/revision/config"

	"github.com/BurntSushi/xgb/xproto"
)

var (
	apps  map[string]*utils.AppCollection
	xutil *utils.X11Interface
)

func main() {
	xutil = utils.CreateNewX11Interface(config.AppCodes)
	apps = make(map[string]*utils.AppCollection)

	availableApps := config.GetAppCodesValues()

	for {
		ev := xutil.WaitForEvent()

		switch event := ev.(type) {
		case xproto.KeyReleaseEvent:
			switch event.Detail {
			case 96:
				{
					printApps()
				}
			case 75:
				fallthrough
			case 76:
				fallthrough
			case 95:
				{
					appName, appExist := config.AppCodes[event.Detail]

					if !appExist {
						log.Printf("App not in config file, please add it first\n")
						continue
					}

					appCollection, collectionExist := apps[appName]
					currentlyFocused := xutil.GetCurrentlyFocused()

					if collectionExist && appCollection.ContainsApp(currentlyFocused) {
						nextWinId := appCollection.GoToNextApp(currentlyFocused)
						xutil.ChangeWindow(nextWinId)
					} else if collectionExist {
						xutil.ChangeWindow(appCollection.Current.WindId)
					} else {
						launchApp(appName)
					}
				}
			}
		case xproto.CreateNotifyEvent:
			{
				className, err := xutil.GetWMClass(event.Window)
				if err != nil {
					for k, v := range apps {
						if v.ContainsApp(xutil.GetCurrentlyFocused()) {
							className = k
						}
					}
				}

				if value, exist := apps[className]; exist {
					value.AddBeforeCurrent(event.Window)
					go runCleanup(apps[className])
				} else if slices.Contains(availableApps, className) {
					apps[className] = utils.CreateNewAppCollection(event.Window)
					go runCleanup(apps[className])
				}

			}
		case xproto.DestroyNotifyEvent:
			{
				for k, v := range apps {
					if v.ContainsApp(event.Window) {
						go deleteKey(v, k)
					}
				}
			}
		}
	}
}

func printApps() {
	if len(apps) == 0 {
		log.Print("No apps currently opened")
	} else {
		for k, v := range apps {
			ids := []xproto.Window{}
			for _, app := range v.AllApps {
				ids = append(ids, app.WindId)
			}
			log.Printf("App: %v, Codes: %X\n", k, ids)
		}
	}
}

func launchApp(appName string) bool {
	err := exec.Command(appName).Start()
	time.Sleep(2 * time.Second)

	if err != nil {
		log.Printf("Failed to launch program: %v", err)
		return false
	}

	return true
}

func deleteKey(collection *utils.AppCollection, key string) {
	time.Sleep(1 * time.Second)
	xutil.CleanUp(collection)
	if collection.Size == 0 {
		delete(apps, key)
	}
}

func runCleanup(appsCollection *utils.AppCollection) {
	time.Sleep(2 * time.Second)
	xutil.CleanUp(appsCollection)
}
