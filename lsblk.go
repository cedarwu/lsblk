package lsblk

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
)

type Device struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	Fsavail    string   `json:"fsavail"`
	Fssize     string   `json:"fssize"`
	Fstype     string   `json:"fstype"`
	Pttype     string   `json:"pttype"`
	Fsused     string   `json:"fsused"`
	Fsuse      string   `json:"fsuse%"`
	Mountpoint string   `json:"mountpoint"`
	Label      string   `json:"label"`
	UUID       string   `json:"uuid"`
	Rm         bool     `json:"rm"`
	Hotplug    bool     `json:"hotplug"`
	Serial     string   `json:"serial"`
	State      string   `json:"state"`
	Group      string   `json:"group"`
	Type       string   `json:"type"`
	Alignment  int      `json:"alignment"`
	Wwn        string   `json:"wwn"`
	Hctl       string   `json:"hctl"`
	Tran       string   `json:"tran"`
	Subsystems string   `json:"subsystems"`
	Rev        string   `json:"rev"`
	Vendor     string   `json:"vendor"`
	Model      string   `json:"model"`
	Children   []Device `json:"children"`
}

func runCmd(command string) (output []byte, err error) {
	if len(command) == 0 {
		return nil, errors.New("invalid command")
	}
	commands := strings.Fields(command)
	output, err = exec.Command(commands[0], commands[1:]...).Output()
	return output, err
}

func PrintDevices(devices map[string]Device) {
	var devList []Device
	for _, dev := range devices {
		devList = append(devList, dev)
	}
	sort.Slice(devList, func(i, j int) bool {
		return devList[i].Name < devList[j].Name
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"name", "hctl", "fstype", "fsavail", "fssize", "fsuse%", "type", "mount", "pttype", "vendor", "model"})

	for _, dev := range devList {
		avail, _ := strconv.ParseUint(dev.Fsavail, 10, 64)
		size, _ := strconv.ParseUint(dev.Fssize, 10, 64)
		table.Append([]string{dev.Name, dev.Hctl, dev.Fstype, humanize.Bytes(avail), humanize.Bytes(size), dev.Fsuse, dev.Type, dev.Mountpoint, dev.Pttype, dev.Vendor, dev.Model})
	}
	table.Render() // Send output
}

func PrintPartitions(devices map[string]Device) {
	partDevMap := make(map[string]string)
	var partList []Device
	for _, dev := range devices {
		for _, child := range dev.Children {
			partDevMap[child.Name] = dev.Name
			child.Vendor = dev.Vendor
			child.Model = dev.Model
			partList = append(partList, child)
		}
	}
	sort.Slice(partList, func(i, j int) bool {
		return partList[i].Name < partList[j].Name
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"disk", "partition", "label", "fstype", "fsavail", "fssize", "fsuse%", "type", "mount", "pttype", "vendor", "model"})

	for _, part := range partList {
		avail, _ := strconv.ParseUint(part.Fsavail, 10, 64)
		size, _ := strconv.ParseUint(part.Fssize, 10, 64)
		table.Append([]string{partDevMap[part.Name], part.Name, part.Label, part.Fstype, humanize.Bytes(avail), humanize.Bytes(size), part.Fsuse, part.Type, part.Mountpoint, part.Pttype, part.Vendor, part.Model})
	}
	table.Render() // Send output
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

	// block, err := ghw.Block()
	// if err != nil {
	// 	return nil, err
	// }

	devices = make(map[string]Device)
	for _, device := range lsblkRsp["blockdevices"] {
		serial, err := getSerial(device.Name)
		if err == nil {
			device.Serial = serial
		}
		devices[device.Name] = device
	}

	return devices, nil
}

func getSerial(devName string) (serial string, err error) {
	output, err := runCmd("bash -c udevadm info --query=property --name=/dev/" + devName + " | grep SCSI_IDENT_SERIAL | awk -F'=' '{print $2}'")
	return string(output), err
}
