package main

import "testing"

func TestAddressResolutionAndSafety(t *testing.T) {
	t.Setenv("PORT", "19666")
	configuration, err := parseConfig(nil)
	if err != nil {
		t.Fatal(err)
	}
	if configuration.Address != "127.0.0.1:19666" {
		t.Fatalf("address=%s", configuration.Address)
	}
	configuration, err = parseConfig([]string{"-addr=127.0.0.1:19777"})
	if err != nil {
		t.Fatal(err)
	}
	if configuration.Address != "127.0.0.1:19777" {
		t.Fatalf("flag did not win: %s", configuration.Address)
	}
	for _, address := range []string{"0.0.0.0:19081", ":19081", "127.0.0.1:80", "example.com:19081"} {
		if err := validateLoopbackAddress(address); err == nil {
			t.Fatalf("unsafe address accepted: %s", address)
		}
	}
}

func TestInvalidPortEnvironment(t *testing.T) {
	t.Setenv("PORT", "not-a-port")
	if _, err := parseConfig(nil); err == nil {
		t.Fatal("invalid PORT accepted")
	}
}
