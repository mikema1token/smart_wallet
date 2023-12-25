package main

import (
	"fmt"
	"testing"
)

func TestPushMessageToTg(t *testing.T) {
	type args struct {
		message string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				message: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PushMessageToTg(tt.args.message)
		})
	}
}

func TestInit(t *testing.T) {
	fmt.Println(config.Test, config.EtherTest, config.EtherMain, config.ChatId, config.TgBotToken, config.HttpProxy)
}
