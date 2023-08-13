package main

import (
	"github.com/getlantern/systray"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func createNewInstance(instancesButton *systray.MenuItem) {
	var mainWindow *walk.MainWindow
	var name *walk.TextEdit
	var path *walk.TextEdit
	var arguments *walk.TextEdit
	var info *walk.TextLabel

	shouldClose := make(chan bool)

	wnd := MainWindow{
		AssignTo: &mainWindow,
		Title:    "SRCDS Manager - New SRCDS Instance",
		Size:     Size{400, 120},
		MinSize:  Size{400, 120},
		MaxSize:  Size{400, 120},
		Layout:   VBox{},
		Children: []Widget{
			TextLabel{
				AssignTo:      &info,
				Text:          "Enter instance details",
				TextColor:     walk.RGB(0, 0, 0),
				TextAlignment: AlignHCenterVCenter,
			},
			TextLabel{Text: "Instance Name"},
			TextEdit{AssignTo: &name},
			TextLabel{Text: "Instance Path"},
			TextEdit{AssignTo: &path},
			TextLabel{Text: "Instance Arguments"},
			TextEdit{AssignTo: &arguments},
			PushButton{
				Text: "Create",
				OnClicked: func() {
					nameText := name.Text()
					pathText := path.Text()
					argumentsText := arguments.Text()

					if nameText == "" || pathText == "" || argumentsText == "" {
						info.SetText("All fields must be filled!")
						info.SetTextColor(walk.RGB(255, 0, 0))
					} else {
						info.SetText("Creating SRCDS instance...")
						info.SetTextColor(walk.RGB(0, 255, 0))
						newInstance := instance{
							Name:      nameText,
							Path:      pathText,
							Arguments: argumentsText,
							State:     INACTIVE,
						}
						instances = append(instances, newInstance)

						addNewInstance(&newInstance, instancesButton)
						shouldClose <- true
					}
				},
			},
		},
	}

	if err := wnd.Create(); err != nil {
		panic(err)
	}

	go func() {
		if <-shouldClose {
			mainWindow.Synchronize(func() {
				mainWindow.Close()
			})
		}
	}()

	mainWindow.Run()

}
