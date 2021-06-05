package lsblk

import "testing"

func TestListDevices(t *testing.T) {
	devices, err := ListDevices()
	if err != nil {
		t.Errorf("list devices failed: %v", err)
	}
	t.Logf("devices: %+v", devices)
}

func TestPrintDevices(t *testing.T) {
	devices, err := ListDevices()
	if err != nil {
		t.Errorf("list devices failed: %v", err)
	}
	PrintDevices(devices)
}

func TestPrintPartitions(t *testing.T) {
	devices, err := ListDevices()
	if err != nil {
		t.Errorf("list devices failed: %v", err)
	}
	PrintPartitions(devices)
}
