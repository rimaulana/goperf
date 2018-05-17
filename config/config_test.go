package config

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func ReaderSuccess(filename string) ([]byte, error) {
	return []byte(filename), nil
}

func ReaderFail(filename string) ([]byte, error) {
	return nil, errors.New(filename)
}

var cases = []struct {
	name   string
	yml    string
	output []ISPConfig
	err    error
	stub   func(filename string) ([]byte, error)
}{
	{
		name: "case success decoding YML",
		yml:  "- name: biznet\n  eth: ens160\n  gateway: 172.16.1.1\n  checkip: 8.8.8.8",
		output: []ISPConfig{{
			Name:    "biznet",
			Eth:     "ens160",
			Gateway: "172.16.1.1",
			CheckIP: "8.8.8.8",
		}},
		err:  nil,
		stub: ReaderSuccess,
	},
	{
		name: "case success decoding YML with multiple data",
		yml:  "- name: biznet\n  eth: ens160\n  gateway: 172.16.1.1\n  checkip: 8.8.8.8\n- name: cbn\n  eth: ens192\n  gateway: 172.16.2.1\n  checkip: 8.8.4.4",
		output: []ISPConfig{
			{
				Name:    "biznet",
				Eth:     "ens160",
				Gateway: "172.16.1.1",
				CheckIP: "8.8.8.8",
			}, {
				Name:    "cbn",
				Eth:     "ens192",
				Gateway: "172.16.2.1",
				CheckIP: "8.8.4.4",
			},
		},
		err:  nil,
		stub: ReaderSuccess,
	},
	{
		name:   "case error in reading source file",
		yml:    "file doesn't existed",
		output: nil,
		err:    errors.New("file doesn't existed"),
		stub:   ReaderFail,
	},
	{
		name:   "case error in parsing yml file",
		yml:    "file doesn't existed",
		output: nil,
		err:    errors.New("yaml: unmarshal errors"),
		stub:   ReaderSuccess,
	},
}

func TestConfig_Load(t *testing.T) {
	for _, tt := range cases {
		cfg := New()
		cfg.StubReader(tt.stub)
		t.Run(tt.name, func(t *testing.T) {
			result, err := cfg.Load(tt.yml)
			if !reflect.DeepEqual(result, tt.output) {
				t.Errorf("Load() = %v, want %v", result, tt.output)
			}
			if err != nil {
				got := err.Error()
				want := tt.err.Error()
				if !strings.Contains(got, want) {
					t.Errorf("Error in Load() = %v, want %v", got, want)
				}
			}
		})
	}
}
