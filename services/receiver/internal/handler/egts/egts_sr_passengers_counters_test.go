package egts

import (
	"bytes"
	"testing"
	// "github.com/google/go-cmp/cmp"
)

var (
	srPassengersCountersBytes    = []byte{0x00, 0x15, 0x14, 0x92, 0x10, 0x00, 0x00, 0x07, 0x03, 0x0A, 0x0F}
	testEgtsSrPassengersCounters = SrPassengersCountersData{
		RawDataFlag:    "0",
		DoorsPresented: "00010101",
		DoorsReleased:  "00010100",
		ModuleAddress:  4242,
		PassengersCountersData: []PassengersCounter{
			{
				DoorNo: 1,
				In:     0,
				Out:    0,
			},
			{
				DoorNo: 3,
				In:     7,
				Out:    3,
			},
			{
				DoorNo: 5,
				In:     10,
				Out:    15,
			},
		},
		PassengersCountersRawData: nil,
	}

	srPassengersCountersRawBytes    = []byte{0x01, 0x15, 0x14, 0x92, 0x10, 0x00, 0x00, 0x07, 0x03, 0x0A, 0x0F}
	testEgtsSrPassengersCountersRaw = SrPassengersCountersData{
		RawDataFlag:               "1",
		DoorsPresented:            "00010101",
		DoorsReleased:             "00010100",
		ModuleAddress:             4242,
		PassengersCountersData:    nil,
		PassengersCountersRawData: []byte{0x00, 0x00, 0x07, 0x03, 0x0A, 0x0F},
	}
)

func TestEgtsSrPassengersCounters_Encode(t *testing.T) {
	passengersCountersDataBytes, err := testEgtsSrPassengersCounters.Encode()
	if err != nil {
		t.Errorf("Ошибка кодирования: %v\n", err)
	}

	if !bytes.Equal(passengersCountersDataBytes, srPassengersCountersBytes) {
		t.Errorf("Байтовые строки не совпадают: %v != %v ", passengersCountersDataBytes, srPassengersCountersBytes)
	}
}

func TestEgtsSrPassengersCounters_EncodeRawCounters(t *testing.T) {
	passengersCountersDataBytes, err := testEgtsSrPassengersCountersRaw.Encode()
	if err != nil {
		t.Errorf("Ошибка кодирования: %v\n", err)
	}

	if !bytes.Equal(passengersCountersDataBytes, srPassengersCountersRawBytes) {
		t.Errorf("Байтовые строки не совпадают: %v != %v ", passengersCountersDataBytes, srPassengersCountersRawBytes)
	}
}

// func TestEgtsSrPassengersCounters_Decode(t *testing.T) {
// 	passengersCountersData := SrPassengersCountersData{}

// 	if err := passengersCountersData.Decode(srPassengersCountersBytes); err != nil {
// 		t.Errorf("Ошибка декодирования: %v\n", err)
// 	}

// 	if diff := cmp.Diff(passengersCountersData, testEgtsSrPassengersCounters); diff != "" {
// 		t.Errorf("Записи не совпадают: (-нужно +сейчас)\n%s", diff)
// 	}
// }

// func TestEgtsSrPassengersCounters_DecodeRawCounters(t *testing.T) {
// 	passengersCountersData := SrPassengersCountersData{}

// 	if err := passengersCountersData.Decode(srPassengersCountersRawBytes); err != nil {
// 		t.Errorf("Ошибка декодирования: %v\n", err)
// 	}

// 	if diff := cmp.Diff(passengersCountersData, testEgtsSrPassengersCountersRaw); diff != "" {
// 		t.Errorf("Записи не совпадают: (-нужно +сейчас)\n%s", diff)
// 	}
// }

// func TestEgtsSrPassengersCounters_Length(t *testing.T) {
// 	if diff := cmp.Diff(uint16(11), testEgtsSrPassengersCounters.Length()); diff != "" {
// 		t.Errorf("Записи не совпадают: (-нужно +сейчас)\n%s", diff)
// 	}
// }

// func TestEgtsSrPassengersCounters_LengthRawCounters(t *testing.T) {
// 	if diff := cmp.Diff(uint16(11), testEgtsSrPassengersCountersRaw.Length()); diff != "" {
// 		t.Errorf("Записи не совпадают: (-нужно +сейчас)\n%s", diff)
// 	}
// }
