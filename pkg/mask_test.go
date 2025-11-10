package protomaskerpkg_test

import (
	"testing"

	. "github.com/bbars/proto-masker/pkg"
	"github.com/stretchr/testify/assert"
)

func TestMaskAsCardPAN(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty",
			s:    "",
			want: "",
		},
		{
			name: "normal",
			s:    "1234567876543210",
			want: "123456******3210",
		},
		{
			name: "spaced",
			s:    "1234 5678 7654 3210",
			want: "1234 56** **** 3210",
		},
		{
			name: "dashed",
			s:    "1234 - 5678 - 7654 - 3210",
			want: "1234 - 56** - **** - 3210",
		},
		{
			name: "short",
			s:    "1234",
			want: "****",
		},
		{
			name: "mid",
			s:    "12345678",
			want: "********",
		},
		{
			name: "long",
			s:    "12345678123456781234567812345678",
			want: "123456**********************5678",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MaskAsCardPAN(tt.s))
		})
	}
}

func TestMaskAsEmail(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty",
			s:    "",
			want: "",
		},
		{
			name: "normal",
			s:    "user@mail.box",
			want: "u***@mail.box",
		},
		{
			name: "short",
			s:    "a@mail.box",
			want: "*@mail.box",
		},
		{
			name: "mid",
			s:    "ann@mail.box",
			want: "***@mail.box",
		},
		{
			name: "long",
			s:    "very-very-long.name+also+has+a+tag@mail.box",
			want: "very-ver**************************@mail.box",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MaskAsEmail(tt.s))
		})
	}
}

func TestMaskAsGeneric(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty",
			s:    "",
			want: "",
		},
		{
			name: "normal",
			s:    "foo bar baz",
			want: "*** *** ***",
		},
		{
			name: "short",
			s:    "foo",
			want: "***",
		},
		{
			name: "mid",
			s:    "boo bar",
			want: "*** ***",
		},
		{
			name: "long",
			s:    "foo bar baz qux fred thud",
			want: "*** *** *** *** **** ****",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MaskAsGeneric(tt.s))
		})
	}
}

func TestMaskAsPassword(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty",
			s:    "",
			want: "",
		},
		{
			name: "normal",
			s:    "qwerty12",
			want: "********",
		},
		{
			name: "short",
			s:    "q",
			want: "********",
		},
		{
			name: "mid",
			s:    "qwer",
			want: "********",
		},
		{
			name: "long",
			s:    "bix$@]L;dXR=tC$Kk3pmSzg2L]giv/9TYs",
			want: "********",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MaskAsPassword(tt.s))
		})
	}
}

func TestMaskAsPhone(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty",
			s:    "",
			want: "",
		},
		{
			name: "normal",
			s:    "+12345678901",
			want: "+1234***8901",
		},
		{
			name: "formatted",
			s:    "+1(234)567-89-01",
			want: "+1(234)***-89-01",
		},
		{
			name: "short",
			s:    "2345",
			want: "****",
		},
		{
			name: "mid",
			s:    "23456789",
			want: "********",
		},
		{
			name: "long",
			s:    "+1 (234-567-89) 012345 67 89 01 23 45 67 89",
			want: "+1 (234-***-**) ****** ** ** ** ** ** 67 89",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MaskAsPhone(tt.s))
		})
	}
}
