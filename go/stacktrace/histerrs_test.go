package stacktrace

import "reflect"
import "testing"

func TestHisterrs(t *testing.T) {
	tests := []struct {
		in   string
		want []*Error
	}{
		{
			in: `
Error detected while processing function Main[2]..<SNR>96_test[1]..<SNR>96_test2[1]..F:
line    3:
E121: Undefined variable: err1`,
			want: []*Error{
				{
					Throwpoint: "function Main[2]..<SNR>96_test[1]..<SNR>96_test2[1]..F[3]",
					Messages:   []string{"E121: Undefined variable: err1"},
				},
			},
		},
		{
			in: `
Error detected while processing function Main[2]..<SNR>96_test[1]..<SNR>96_test2[1]..F:
line    3:
E121: Undefined variable: err1
E15: Invalid expression: err1
line    4:
E121: Undefined variable: err2
E15: Invalid expression: err2
Error detected while processing /path/to/file.vim:
line   33:
E605: Exception not caught: 0
			`,
			want: []*Error{
				{
					Throwpoint: "function Main[2]..<SNR>96_test[1]..<SNR>96_test2[1]..F[3]",
					Messages: []string{
						"E121: Undefined variable: err1",
						"E15: Invalid expression: err1",
					},
				},
				{
					Throwpoint: "function Main[2]..<SNR>96_test[1]..<SNR>96_test2[1]..F[4]",
					Messages: []string{
						"E121: Undefined variable: err2",
						"E15: Invalid expression: err2",
					},
				},
				{
					Throwpoint: "/path/to/file.vim[33]",
					Messages: []string{
						"E605: Exception not caught: 0",
					},
				},
			},
		},
		{
			in: `
ok1
Error detected while processing function F1:
line    3:
E121: errormsg
ignore
Error detected while processing function G1:
invalid!
line    3:
E121: errormsg
ok2 after invalid
Error detected while processing function F2:
line    3:
E121: errormsg
Error detected while processing function G2:
line    3:
invalid
E121: errormsg
Error detected while processing function F3:
line    3:
E121: errormsg
line    4:
invalid
E121: errormsg
`,
			want: []*Error{
				{Throwpoint: "function F1[3]", Messages: []string{"E121: errormsg"}},
				{Throwpoint: "function F2[3]", Messages: []string{"E121: errormsg"}},
				{Throwpoint: "function F3[3]", Messages: []string{"E121: errormsg"}},
			},
		},
	}
	v := &Vim{c: cli}
	for _, tt := range tests {
		got, err := Histerrs(tt.in)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("in:\n%v", tt.in)
			t.Log("got:")
			for _, e := range got {
				t.Logf("%#v", e)
			}
			t.Log("want:")
			for _, e := range tt.want {
				t.Logf("%#v", e)
			}
		}
		// check all throwpints are valid
		for _, e := range got {
			ss, err := v.Build(e.Throwpoint)
			if err != nil {
				t.Errorf("Error.Throwpoint (%v) is invalid: %v", e.Throwpoint, err)
			}
			if len(ss.Stacks) == 0 {
				t.Errorf("Error.Throwpoint is empty: Error: %v", e)
			}
		}
	}
}
