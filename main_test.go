package main

import (
	"testing"
	"time"
)

func Test_durationFormatter(t *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantRes string
	}{
		{
			name:    "hour+minute+second",
			args:    args{d: time.Hour*2 + time.Minute*30 + time.Second*40},
			wantRes: "02h:30m:40s",
		},
		{
			name:    "hour+second",
			args:    args{d: time.Hour + time.Second*25},
			wantRes: "01h:00m:25s",
		},
		{
			name:    "minute+second",
			args:    args{d: time.Minute*59 + time.Second*40},
			wantRes: "59m:40s",
		},
		{
			name:    "second",
			args:    args{d: time.Second * 2},
			wantRes: "02s",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := durationFormatter(tt.args.d); gotRes != tt.wantRes {
				t.Errorf("durationFormatter() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
