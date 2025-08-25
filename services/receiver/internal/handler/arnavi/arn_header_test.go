package arnavi

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// После выполнения тестов, если нужно, можно просмотреть результаты покрытия
// тестами, выполнив go test -cover
// go test -v для более детального вывода.

var (
	bytesHeadone = []byte{0xFF, 0x23, 0xE9, 0xEF, 0x78, 0x2D, 0xE7, 0x12, 0x03, 0x00}
	testHeadOne  = HeadOne{
		Signature: 0xFF,
		Version:   0x23,
		IdImei:    865209039777769,
	}
)

func TestDumpHeadOne_Encode(t *testing.T) {
	countersBytes, err := testHeadOne.Encode()
	if assert.NoError(t, err) {
		assert.Equal(t, countersBytes, bytesHeadone)
	}
}

func TestDumpHeadOne_decode(t *testing.T) {
	newHead := HeadOne{}
	err := newHead.Decode(bytesHeadone)
	if assert.NoError(t, err) {
		assert.Equal(t, newHead, testHeadOne)
	}
}
func TestHeadOne_Length(t *testing.T) {
	l := testHeadOne.Length()
	fmt.Println("actual len:", l)
	assert.Equal(t, int(l), len(bytesHeadone))
}

func TestHeadOne_Decode(t *testing.T) {
	type fields struct {
		Signature byte
		Version   byte
		IdImei    uint64
		ExtId     uint64
	}
	type args struct {
		content []byte
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
			e := &HeadOne{
				Signature: tt.fields.Signature,
				Version:   tt.fields.Version,
				IdImei:    tt.fields.IdImei,
				ExtId:     tt.fields.ExtId,
			}
			if err := e.Decode(tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("HeadOne.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHeadOne_Encode(t *testing.T) {
	type fields struct {
		Signature byte
		Version   byte
		IdImei    uint64
		ExtId     uint64
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
			e := &HeadOne{
				Signature: tt.fields.Signature,
				Version:   tt.fields.Version,
				IdImei:    tt.fields.IdImei,
				ExtId:     tt.fields.ExtId,
			}
			got, err := e.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("HeadOne.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HeadOne.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}
