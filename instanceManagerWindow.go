package main

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"log"
)

type MainWindowType struct {
	*walk.MainWindow
	ListBox          *walk.ListBox
	ListBoxComposite *walk.Composite
	InfoLabel        *walk.TextLabel
	NameEntry        *walk.TextEdit
	PathEntry        *walk.TextEdit
	ArgsEntry        *walk.TextEdit
}

func openSRCDSManagerMenu() {
	mw := &MainWindowType{}

	names := make([]string, len(instances))
	for i, inst := range instances {
		names[i] = inst.Name
	}

	shouldClose := make(chan bool)

	if _, err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "SRCDS Manager - New SRCDS Instance",
		Size:     Size{Width: 500, Height: 200},
		MinSize:  Size{Width: 500, Height: 200},
		Layout:   HBox{},
		Children: []Widget{
			Composite{
				AssignTo: &mw.ListBoxComposite,
				Layout:   VBox{MarginsZero: true},
				Children: []Widget{
					ListBox{
						AssignTo: &mw.ListBox,
						MaxSize:  Size{Width: 100},
						Model:    names,
						OnCurrentIndexChanged: func() {
							currentIndex := mw.ListBox.CurrentIndex()
							if currentIndex >= 0 && currentIndex < len(instances) {
								selectedInstance := instances[currentIndex]
								_ = mw.NameEntry.SetText(selectedInstance.Name)
								_ = mw.PathEntry.SetText(selectedInstance.Path)
								_ = mw.ArgsEntry.SetText(selectedInstance.Arguments)
							}
						},
					},
					PushButton{
						Text: "Create New",
						OnClicked: func() {
							defaultInstance := instance{
								Name:      "New Instance Name",
								Path:      "SRCDS Path",
								Arguments: "SRCDS Launch Arguments",
							}

							instances = append(instances, defaultInstance)

							names = append(names, defaultInstance.Name)
							_ = mw.ListBox.SetCurrentIndex(len(names))

							_ = mw.ListBox.SetModel(names)
						},
					},
				},
			},
			Composite{
				Layout: VBox{},
				Children: []Widget{
					TextLabel{
						AssignTo:  &mw.InfoLabel,
						Text:      "Enter instance details",
						TextColor: walk.RGB(0, 0, 0),
					},
					TextLabel{Text: "Instance Name"},
					TextEdit{AssignTo: &mw.NameEntry},
					TextLabel{Text: "Instance Path"},
					TextEdit{AssignTo: &mw.PathEntry},
					TextLabel{Text: "Instance Arguments"},
					TextEdit{AssignTo: &mw.ArgsEntry},
					PushButton{
						Text: "Save",
						OnClicked: func() {
							nameText := mw.NameEntry.Text()
							pathText := mw.PathEntry.Text()
							argumentsText := mw.ArgsEntry.Text()

							if nameText == "" || pathText == "" || argumentsText == "" {
								_ = mw.InfoLabel.SetText("All fields must be filled!")
								mw.InfoLabel.SetTextColor(walk.RGB(255, 0, 0))
							} else {
								_ = mw.InfoLabel.SetText("Saving SRCDS instance...")
								mw.InfoLabel.SetTextColor(walk.RGB(0, 255, 0))

								oldInstances := instances

								clearInstancesTray()
								instances = make([]instance, 0)

								for i, inst := range oldInstances {
									if i == mw.ListBox.CurrentIndex() {
										instances = append(instances, instance{
											Name:      nameText,
											Path:      pathText,
											Arguments: argumentsText,
										})
										fmt.Println("Saved instance: " + nameText)
									} else {
										instances = append(instances, inst)
										fmt.Println("Created instance: " + inst.Name)
									}
								}

								names = make([]string, len(instances))
								for i, inst := range instances {
									names[i] = inst.Name
								}
								_ = mw.ListBox.SetModel(names)

								populateInstancesTray()
							}
						},
					},
				},
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}

	go func() {
		if <-shouldClose {
			mw.MainWindow.Synchronize(func() {
				err := mw.MainWindow.Close()
				if err != nil {
					return
				}
			})
		}
	}()
}
