package lsblk

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/jinzhu/copier"
	"github.com/olekukonko/tablewriter"
)

type Device struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	Fsavail    uint64   `json:"fsavail"`
	Fssize     uint64   `json:"fssize"`
	Fsused     uint64   `json:"fsused"`
	Fsusage    uint     `json:"fsusage"` // percent that was used
	Fstype     string   `json:"fstype"`
	Pttype     string   `json:"pttype"`
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

type _Device struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Fsavail    string    `json:"fsavail"`
	Fssize     string    `json:"fssize"`
	Fstype     string    `json:"fstype"`
	Pttype     string    `json:"pttype"`
	Fsused     string    `json:"fsused"`
	Fsuse      string    `json:"fsuse%"`
	Mountpoint string    `json:"mountpoint"`
	Label      string    `json:"label"`
	UUID       string    `json:"uuid"`
	Rm         bool      `json:"rm"`
	Hotplug    bool      `json:"hotplug"`
	Serial     string    `json:"serial"`
	State      string    `json:"state"`
	Group      string    `json:"group"`
	Type       string    `json:"type"`
	Alignment  int       `json:"alignment"`
	Wwn        string    `json:"wwn"`
	Hctl       string    `json:"hctl"`
	Tran       string    `json:"tran"`
	Subsystems string    `json:"subsystems"`
	Rev        string    `json:"rev"`
	Vendor     string    `json:"vendor"`
	Model      string    `json:"model"`
	Children   []_Device `json:"children"`
}

func runCmd(command string) (output []byte, err error) {
	if len(command) == 0 {
		return nil, errors.New("invalid command")
	}
	commands := strings.Fields(command)
	output, err = exec.Command(commands[0], commands[1:]...).Output()
	return output, err
}

func runBash(command string) (output []byte, err error) {
	if len(command) == 0 {
		return nil, errors.New("invalid command")
	}
	output, err = exec.Command("bash", "-c", command).Output()
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
	table.SetHeader([]string{"name", "hctl", "fstype", "fssize", "fsused", "fsavail", "fsuse%", "type", "mount", "pttype", "vendor", "model"})

	for _, dev := range devList {
		table.Append([]string{dev.Name, dev.Hctl, dev.Fstype, humanize.Bytes(dev.Fssize), humanize.Bytes(dev.Fsused), humanize.Bytes(dev.Fsavail), strconv.FormatUint(uint64(dev.Fsusage), 10) + "%", dev.Type, dev.Mountpoint, dev.Pttype, dev.Vendor, dev.Model})
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
	table.SetHeader([]string{"disk", "partition", "label", "fstype", "fssize", "fsused", "fsavail", "fsuse%", "type", "mount", "pttype", "vendor", "model"})

	for _, part := range partList {
		table.Append([]string{partDevMap[part.Name], part.Name, part.Label, part.Fstype, humanize.Bytes(part.Fssize), humanize.Bytes(part.Fsused), humanize.Bytes(part.Fsavail), strconv.FormatUint(uint64(part.Fsusage), 10) + "%", part.Type, part.Mountpoint, part.Pttype, part.Vendor, part.Model})
	}
	table.Render() // Send output
}

// NewLSSCSI is a constructor for LSSCSI
func ListDevices() (devices map[string]Device, err error) {
	output, err := runCmd("lsblk -e7 -b -J -o name,path,fsavail,fssize,fstype,pttype,fsused,fsuse%,mountpoint,label,uuid,rm,hotplug,serial,state,group,type,alignment,wwn,hctl,tran,subsystems,rev,vendor,model")
	if err != nil {
		return nil, err
	}

	lsblkRsp := make(map[string][]_Device)
	err = json.Unmarshal(output, &lsblkRsp)
	if err != nil {
		return nil, err
	}

	devices = make(map[string]Device)
	for _, _device := range lsblkRsp["blockdevices"] {
		var device Device
		copier.Copy(&device, &_device)

		device.Fsavail, _ = strconv.ParseUint(_device.Fsavail, 10, 64)
		device.Fsused, _ = strconv.ParseUint(_device.Fsused, 10, 64)
		device.Fssize, _ = strconv.ParseUint(_device.Fssize, 10, 64)
		if device.Fssize > 0 {
			device.Fsusage = uint(math.Round(float64(device.Fsused*100) / float64(device.Fssize)))
		}

		for i, child := range _device.Children {
			device.Children[i].Fsavail, _ = strconv.ParseUint(child.Fsavail, 10, 64)
			device.Children[i].Fsused, _ = strconv.ParseUint(child.Fsused, 10, 64)
			device.Children[i].Fssize, _ = strconv.ParseUint(child.Fssize, 10, 64)
			if device.Children[i].Fssize > 0 {
				device.Children[i].Fsusage = uint(math.Round(float64(device.Children[i].Fsused*100) / float64(device.Children[i].Fssize)))
			}
		}

		serial, err := getSerial(device.Name)
		if err == nil {
			device.Serial = serial
		}
		devices[device.Name] = device
	}

	return devices, nil
}

func getSerial(devName string) (serial string, err error) {
	output, err := runBash("udevadm info --query=property --name=/dev/" + devName + " | grep SCSI_IDENT_SERIAL | awk -F'=' '{print $2}'")
	return strings.TrimSpace(string(output)), err
}
