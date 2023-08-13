package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	INACTIVE int = 0
	RUNNING      = 1
	ERRORED      = 2
)

type instance struct {
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	Arguments    string            `json:"arguments"`
	Branch       string            `json:"branch"`
	State        int               `json:"-"`
	CMD          *exec.Cmd         `json:"-"`
	Button       *systray.MenuItem `json:"-"`
	DeleteButton *systray.MenuItem `json:"-"`
	EditButton   *systray.MenuItem `json:"-"`
}

var instances []instance

func main() {
	fmt.Println("Ready.")

	instances = make([]instance, 0)

	instances = loadInstances()
	systray.Run(onReady, onExit)
}

const fileName = "instances.json"

func saveAllInstances() error {
	jsonData, err := json.Marshal(instances)
	if err != nil {
		panic(err)
	}

	return os.WriteFile(fileName, jsonData, 0777)
}

func loadInstances() []instance {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}

	}(file)

	data, err := io.ReadAll(file)

	if err != nil {
		panic(err)
	}

	var loadedInstances []instance
	err = json.Unmarshal(data, &loadedInstances)
	if err != nil {
		return nil
	}

	return loadedInstances
}

//go:embed icons/icon.ico
var icon []byte

//go:embed icons/icon.png
var _ []byte

//go:embed icons/srcds.ico
var srcdsIcon []byte

//go:embed icons/srcdsx64.ico
var srcdsx64Icon []byte

func clearInstancesTray() {
	for _, trayItems := range instances {
		if trayItems.Button != nil {
			trayItems.Button.Hide()
		}
	}
}

var instancesButton *systray.MenuItem

func populateInstancesTray() {
	clearInstancesTray()

	for i := range instances {
		inst := &instances[i]

		srcdsInstanceButton := instancesButton.AddSubMenuItem(inst.Name, inst.Path)
		if strings.Contains(inst.Path, "64") {
			inst.Branch = "x86-64"
		}
		if inst.Branch == "x86-64" {
			srcdsInstanceButton.SetIcon(srcdsx64Icon)
			srcdsInstanceButton.SetTitle(inst.Name + " (x86-64)")
		} else {
			srcdsInstanceButton.SetIcon(srcdsIcon)
		}

		inst.Button = srcdsInstanceButton

		go listenButton(inst)
	}

	err := saveAllInstances()
	if err != nil {
		return
	}
}

func listenButton(inst *instance) {
	for range inst.Button.ClickedCh {
		if inst.State == INACTIVE {
			startInstance(inst)
		} else if inst.State == RUNNING {
			stopInstance(inst)
		}
	}
}

func onReady() {
	systray.SetIcon(icon)
	systray.SetTitle("SRCDS Manager")
	systray.SetTooltip("SRCDS Manager")

	instancesButton = systray.AddMenuItem("Instances", "")

	openMenu := systray.AddMenuItem("Manage Instances", "")

	systray.AddSeparator()

	populateInstancesTray()

	exit := systray.AddMenuItem("Exit", "Exit the app")

	go func() {
		for {
			select {
			case <-openMenu.ClickedCh:
				fmt.Println("Opening SRCDS Manager Menu")
				openSRCDSManagerMenu()
			case <-exit.ClickedCh:
				systray.Quit()
				os.Exit(0)
				return
			}
		}
	}()
}

func onExit() {
}

func startInstance(instance *instance) {
	fmt.Println("Starting SRCDS instance: " + instance.Name)
	instance.State = RUNNING

	var title = instance.Name
	var oldTitle string

	if instance.Branch == "x86-64" {
		title = title + " (x86-64)"
	}
	oldTitle = title
	title = title + " [RUNNING]"

	instance.Button.SetTitle(title)

	cmd := exec.Command(instance.Path, instance.Arguments)
	instance.CMD = cmd

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Start()

	if err != nil {
		err := beeep.Notify("Error starting SRCDS instance!", "There was an error starting the '"+instance.Name+"' SRCDS instance!", "icons/icon.png")
		if err != nil {
			return
		}
		instance.State = ERRORED
		instance.Button.SetTitle(oldTitle + " [ERRORED]")
		return
	}

	go func() {
		err = cmd.Wait()
		if err != nil {
			var exitError *exec.ExitError
			ok := errors.As(err, &exitError)

			if ok {
				fmt.Println("SRCDS Instance '" + instance.Name + "' exited OK.")
				instance.State = INACTIVE
				instance.Button.SetTitle(oldTitle)
			} else {
				fmt.Println("SRCDS Instance '" + instance.Name + "' exited BADLY.")
				fmt.Println(err)
				err := beeep.Notify("Error while exiting SRCDS instance!", "There was an error while exiting the '"+instance.Name+"' SRCDS instance!", "icons/icon.png")
				if err != nil {
					return
				}
				instance.State = ERRORED
				instance.Button.SetTitle(oldTitle + " [ERRORED]")
			}
		} else {
			fmt.Println("SRCDS instance '" + instance.Name + "' closed")
			instance.State = INACTIVE
			instance.Button.SetTitle(oldTitle)
			startInstance(instance)
		}
	}()
}

func stopInstance(instance *instance) {
	fmt.Println("Stopping SRCDS instance: " + instance.Name)
	instance.State = RUNNING
	err := instance.CMD.Process.Kill()
	if err != nil {
		return
	}
}
