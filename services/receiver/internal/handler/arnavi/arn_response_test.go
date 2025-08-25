package arnavi

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	bytesAnswer = []byte{0x7B, 0x04, 0x00, 0x7F, 0x21, 0x5A, 0xAA, 0x5A, 0x7D}
	data        = ConfirmationHeader{
		TimeAnswer: 1521113633,
	}
	testAnswer = ResCom{
		StartSign: 0x7B,
		Size:      4,
		CodeCom:   0,
		CheckSum:  0x7F,
		Data:      &data,
		EndSign:   0x7D,
	}
)

func TestResComDump_Decode(t *testing.T) {
	newAnswer := ResCom{}
	err := newAnswer.Decode(bytesAnswer)
	if assert.NoError(t, err) {
		assert.Equal(t, newAnswer, testAnswer)
	}
}
func TestResComDump_Encode(t *testing.T) {
	// fmt.Println("size: ", data.Length())
	buteTest, err := testAnswer.Encode()
	if assert.NoError(t, err) {
		assert.Equal(t, bytesAnswer, buteTest)
	}
}

func TestResCom_Decode(t *testing.T) {
	type fields struct {
		StartSign byte
		Size      byte
		CodeCom   byte
		CheckSum  byte
		Data      BinaryData
		EndSign   byte
	}
	type args struct {
		rec []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResCom{
				StartSign: tt.fields.StartSign,
				Size:      tt.fields.Size,
				CodeCom:   tt.fields.CodeCom,
				CheckSum:  tt.fields.CheckSum,
				Data:      tt.fields.Data,
				EndSign:   tt.fields.EndSign,
			}
			if err := r.Decode(tt.args.rec); (err != nil) != tt.wantErr {
				t.Errorf("ResCom.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResCom_Encode(t *testing.T) {
	type fields struct {
		StartSign byte
		Size      byte
		CodeCom   byte
		CheckSum  byte
		Data      BinaryData
		EndSign   byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResCom{
				StartSign: tt.fields.StartSign,
				Size:      tt.fields.Size,
				CodeCom:   tt.fields.CodeCom,
				CheckSum:  tt.fields.CheckSum,
				Data:      tt.fields.Data,
				EndSign:   tt.fields.EndSign,
			}
			got, err := r.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("ResCom.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResCom.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}
