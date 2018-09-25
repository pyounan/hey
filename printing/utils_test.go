package printing

import (
	_ "image/png"
	"pos-proxy/config"
	"testing"
)

func TestSetLang(t *testing.T) {
	type args struct {
		rcrs      string
		fdmConfig []config.FDMConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "No FDM", args: args{rcrs: ""}, want: "en-us"},
		{args: args{rcrs: "x", fdmConfig: []config.FDMConfig{config.FDMConfig{RCRS: "x", Language: "ar-eg"}}}, want: "ar-eg"},
	}
	for _, tt := range tests {
		config.Config.IsFDMEnabled = true
		config.Config.FDMs = tt.args.fdmConfig
		t.Run(tt.name, func(t *testing.T) {
			if got := SetLang(tt.args.rcrs); got != tt.want {
				t.Errorf("SetLang() = %v, want %v", got, tt.want)
			}
		})
	}
}
