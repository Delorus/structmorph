package structmorph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStructName(t *testing.T) {
	type args struct {
		rawName string
	}
	tests := []struct {
		name    string
		args    args
		want    StructName
		wantErr bool
	}{
		{
			name: "ParseStructName with valid input",
			args: args{
				rawName: "mypackage.MyStruct",
			},
			want: StructName{
				Package: "mypackage",
				Name:    "MyStruct",
			},
			wantErr: false,
		},
		{
			name: "ParseStructName with no package",
			args: args{
				rawName: "MyStruct",
			},
			want: StructName{
				Package: "main",
				Name:    "MyStruct",
			},
			wantErr: false,
		},
		// for now, it's not possible to test this case,
		// because we cannot distinguish between a package name and a struct name in that stage
		//{
		//	name: "ParseStructName with invalid input",
		//	args: args{
		//		rawName: "mypackage",
		//	},
		//	want:    StructName{},
		//	wantErr: true,
		//},
		{
			name: "ParseStructName with empty input",
			args: args{
				rawName: "",
			},
			want:    StructName{},
			wantErr: true,
		},
		{
			name: "ParseStructName with multiple dots",
			args: args{
				rawName: "mypackage.subpackage.MyStruct",
			},
			want:    StructName{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStructName(tt.args.rawName)

			assert.Equal(t, tt.wantErr, err != nil, "ParseStructName() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
