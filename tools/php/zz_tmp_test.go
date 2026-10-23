package php

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestTmpStart(t *testing.T) {
	ver := "7.4.33"
	if err := InstallVersion(ver, true); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(4 * time.Minute)
	for {
		vers, _ := ListVersions()
		found := false
		for _, v := range vers {
			if v.Version == ver {
				found = true
			}
		}
		if found {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("install timeout")
		}
		time.Sleep(time.Second)
	}
	defer RemoveVersion(ver)

	cfg := InstanceConfig{Name: "test-site", Version: ver, Port: 18080, NTS: true}
	if err := SaveInstance(cfg); err != nil {
		t.Fatal(err)
	}
	list, _ := ListInstances()
	id := list[len(list)-1].ID
	defer RemoveInstance(id)

	if err := StartInstance(id); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	st, _ := GetInstanceStatus(id)
	fmt.Println("running:", st.Running, "pid:", st.PID)
	fmt.Println("configPath:", st.ConfigPath)
	fmt.Println("docRoot:", st.DocRoot)
	if err := RestartInstance(id); err != nil {
		t.Fatalf("restart failed: %v", err)
	}
	fmt.Println("restart ok")
	logs, _ := Logs(id, 10)
	for _, l := range logs {
		fmt.Println("log:", l)
	}
	if err := StopInstance(id); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	fmt.Println("stop ok")
	_ = os.RemoveAll(st.DocRoot)
}
