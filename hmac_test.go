package main

import (
	"encoding/base64"
	"testing"
)

const Base64EncodedHmacSecret = "p0dvfvuavPGx5vvQTsimYQ=="
const Base64EncodedIsReadyHmac = "xzLggfnMn5hWXlUwqBWAESCDOQMGRgBGeXgBJunhpOY="

func TestCreateHmac(t *testing.T) {
	secretBytes, err := base64.StdEncoding.DecodeString(Base64EncodedHmacSecret)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	rslt := CreateHmac([]byte("GET"), secretBytes)
	actualHmac := base64.StdEncoding.EncodeToString(rslt)
	if actualHmac != "Jy+DKLmNaARI2fZkfut1LZW15Cvyllt1pLjRATdIT+I=" {
		t.Logf("Expected Hmac: %s but received: %s\n", Base64EncodedIsReadyHmac, actualHmac)
		t.Fail()
	}
}

func TestCreateRequestHmac(t *testing.T) {
	rslt := CreateRequestHmac("GET", "/ready", "", Base64EncodedHmacSecret)
	actualHmac := base64.StdEncoding.EncodeToString(rslt)
	if actualHmac != Base64EncodedIsReadyHmac {
		t.Logf("Expected Hmac: %s but received: %s\n", Base64EncodedIsReadyHmac, actualHmac)
		t.Fail()
	}
}

func TestIsReadyHmac(t *testing.T) {
	UpdateCurrentSettings(DefaultSettings())
	SetHmacSecret(Base64EncodedHmacSecret)
	SetCurrentPort("0")
	req, err := CreateRequestForGetHandler("ready")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	actualHmac := req.Header.Get(HmacHeaderName)
	if actualHmac != Base64EncodedIsReadyHmac {
		t.Logf("Expected Hmac: %s but received: %s\n", Base64EncodedIsReadyHmac, actualHmac)
		t.Fail()
	}
}
