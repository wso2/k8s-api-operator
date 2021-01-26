package registry

import (
	"testing"
)

func TestGetPathWithoutReg(t *testing.T) {
	datas :=
		[]struct {
			in  string
			out string
		}{
			{
				in:  "192.168.8.132:5000/apis-test",
				out: "apis-test",
			},
			{
				in:  "192.168.8.132:5000",
				out: "",
			},
		}

	for _, data := range datas {
		result := getPathWithoutReg(data.in)
		if result != data.out {
			t.Errorf("Registry image path without registry host is not matching, getPathWithoutReg(%v) = %v, want = %v", data.in, result, data.out)
		}
	}
}
