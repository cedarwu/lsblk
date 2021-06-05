package lsblk

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

type Device struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Fsavail    string `json:"fsavail"`
	Fssize     string `json:"fssize"`
	Fstype     string `json:"fstype"`
	Pttype     string `json:"pttype"`
	Fsused     string `json:"fsused"`
	Fsuse      string `json:"fsuse%"`
	Mountpoint string `json:"mountpoint"`
	Label      string `json:"label"`
	UUID       string `json:"uuid"`
	Rm         bool   `json:"rm"`
	Hotplug    bool   `json:"hotplug"`
	Serial     string `json:"serial"`
	State      string `json:"state"`
	Group      string `json:"group"`
	Type       string `json:"type"`
	Alignment  int    `json:"alignment"`
	Wwn        string `json:"wwn"`
	Hctl       string `json:"hctl"`
	Tran       string `json:"tran"`
	Subsystems string `json:"subsystems"`
	Rev        string `json:"rev"`
	Vendor     string `json:"vendor"`
	Model      string `json:"model"`
}

func runCmd(command string) (output []byte, err error) {
	if len(command) == 0 {
		return nil, errors.New("invalid command")
	}
	commands := strings.Fields(command)
	output, err = exec.Command(commands[0], commands[1:]...).Output()
	return output, err
}

// NewLSSCSI is a constructor for LSSCSI
func ListDevices() (devices map[string]Device, err error) {
	output, err := runCmd("lsblk -e7 -b -J -o name,path,fsavail,fssize,fstype,pttype,fsused,fsuse%,mountpoint,label,uuid,rm,hotplug,serial,state,group,type,alignment,wwn,hctl,tran,subsystems,rev,vendor,model")
	if err != nil {
		return nil, err
	}

	lsblkRsp := make(map[string][]Device)
	err = json.Unmarshal(output, &lsblkRsp)
	if err != nil {
		return nil, err
	}

	devices = make(map[string]Device)
	for _, device := range lsblkRsp["blockdevices"] {
		devices[device.Name] = device
	}
	return devices, nil
}
