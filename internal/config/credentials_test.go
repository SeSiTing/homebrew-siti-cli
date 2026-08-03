package config

import (
	"testing"

	keyring "github.com/zalando/go-keyring"
)

func TestCredentialLifecycle(t *testing.T) {
	keyring.MockInit()
	if err := SetCredential("BAILIAN", "secret-token"); err != nil {
		t.Fatal(err)
	}
	ok, err := HasCredential("bailian")
	if err != nil || !ok {
		t.Fatalf("HasCredential=%v err=%v", ok, err)
	}
	token, err := GetCredential("bailian")
	if err != nil || token != "secret-token" {
		t.Fatalf("GetCredential=%q err=%v", token, err)
	}
	if err := DeleteCredential("bailian"); err != nil {
		t.Fatal(err)
	}
	if ok, err := HasCredential("bailian"); err != nil || ok {
		t.Fatalf("after delete HasCredential=%v err=%v", ok, err)
	}
}
