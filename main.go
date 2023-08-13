package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
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
	Name      string            `json:"name"`
	Path      string            `json:"path"`
	Arguments string            `json:"arguments"`
	Branch    string            `json:"branch"`
	Button    *systray.MenuItem `json:"-"`
	State     int               `json:"-"`
	CMD       *exec.Cmd         `json:"-"`
}

type globalArgument struct {
	argument string
}

var instances []instance

func main() {
	fmt.Println("Ready.")

	instances = make([]instance, 0)
	globalArguments = make([]globalArgument, 0)

	instances, _ = loadInstances()
	systray.Run(onReady, onExit)
}

const fileName = "instances.json"

func saveAllInstances() error {
	jsonData, err := json.Marshal(instances)
	if err != nil {
		panic(err)
	}

	return os.WriteFile(fileName, []byte(jsonData), 0777)
}

func loadInstances() ([]instance, error) {
	data, err := os.ReadFile(fileName)

	if err != nil {
		panic(err)
	}

	var loadedInstances []instance
	err = json.Unmarshal(data, &loadedInstances)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling instances: %w", err)
	}

	return loadedInstances, nil
}

//go:embed icons/icon.ico
var icon []byte

//go:embed icons/icon.png
var _ []byte

//go:embed icons/srcds.ico
var srcdsIcon []byte

//go:embed icons/srcdsx64.ico
var srcdsx64Icon []byte

func addNewInstance(srcdsInst *instance, instancesButton *systray.MenuItem) {
	srcdsInstanceButton := instancesButton.AddSubMenuItem(srcdsInst.Name, srcdsInst.Path)
	if strings.Contains(srcdsInst.Path, "64") {
		srcdsInst.Branch = "x86-64"
	}
	if srcdsInst.Branch == "x86-64" {
		srcdsInstanceButton.SetIcon(srcdsx64Icon)
		srcdsInstanceButton.SetTitle(srcdsInst.Name + " (x86-64)")
	} else {
		srcdsInstanceButton.SetIcon(srcdsIcon)
	}

	srcdsInst.Button = srcdsInstanceButton

	go func(btn *systray.MenuItem, inst *instance) {
		for {
			<-btn.ClickedCh
			if inst.State == INACTIVE {
				startInstance(inst)
			} else if inst.State == RUNNING {
				stopInstance(inst)
			}
		}
	}(srcdsInstanceButton, srcdsInst)
	err := saveAllInstances()
	if err != nil {
		return
	}
}

func onReady() {
	systray.SetIcon(icon)
	systray.SetTitle("Fuck")
	systray.SetTooltip("tooltip")

	// Instances

	instancesButton := systray.AddMenuItem("Instances", "")
	addNewSRCDSInstanceButton := instancesButton.AddSubMenuItem("Add New", "Add new SRCDS instance.")

	systray.AddSeparator()

	for i := range instances {
		addNewInstance(&instances[i], instancesButton)
	}

	// exit
	exit := systray.AddMenuItem("Exit", "Exit the app")

	go func() {
		for {
			select {
			case <-instancesButton.ClickedCh:
			case <-addNewSRCDSInstanceButton.ClickedCh:
				fmt.Println("Opening SRCDS Manager Menu")
				createNewInstance(instancesButton)
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
			_, ok := err.(*exec.ExitError)

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
