package lsblk

import "testing"

func TestListDevices(t *testing.T) {
	devices, err := ListDevices()
	if err != nil {
		t.Errorf("list devices failed: %v", err)
	}
	t.Logf("devices: %+v", devices)
}
