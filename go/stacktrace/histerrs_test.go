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
		got := Histerrs(tt.in)
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

func TestSelectError(t *testing.T) {
	msghist := `
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
`

	defer func(f func(cli *Vim, candidates []string) (int, error)) {
		inputlist = f
	}(inputlist)

	selected := 0

	inputlist = func(_ *Vim, _ []string) (int, error) {
		return selected, nil
	}

	tests := []struct {
		selected       int
		wantThrowpoint string
		wantNil        bool
		wantErr        bool
	}{
		{selected: -1, wantErr: true},
		{selected: 0, wantNil: true},
		{selected: 1, wantThrowpoint: "function Main[2]..<SNR>96_test[1]..<SNR>96_test2[1]..F[3]"},
		{selected: 2, wantThrowpoint: "function Main[2]..<SNR>96_test[1]..<SNR>96_test2[1]..F[4]"},
		{selected: 3, wantThrowpoint: "/path/to/file.vim[33]"},
		{selected: 4, wantErr: true},
	}
	v := &Vim{c: cli}
	for _, tt := range tests {
		selected = tt.selected
		got, err := v.selectError(msghist)
		if err != nil {
			if !tt.wantErr {
				t.Error(err)
			}
			continue
		}
		if tt.wantNil {
			if got != nil {
				t.Errorf("want nil but got %v", got)
			}
			continue
		}
		if got.Throwpoint != tt.wantThrowpoint {
			t.Errorf("got %v, want %v", got.Throwpoint, tt.wantThrowpoint)
		}
	}
}

func TestVimFromhist(t *testing.T) {
	v := &Vim{c: cli}
	got, err := v.Fromhist()
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("Vim.Fromhist() = %v, want nil", got)
	}
}

func TestSelectError_empty(t *testing.T) {
	v := &Vim{c: cli}
	got, err := v.selectError("")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("Vim.selectError('') = %v, want nil", got)
	}
}

func TestSelectError_one(t *testing.T) {
	msghist := `
Error detected while processing function Main[2]..<SNR>96_test[1]..<SNR>96_test2[1]..F:
line    3:
E121: Undefined variable: err1
E15: Invalid expression: err1
`
	wantThrowpoint := "function Main[2]..<SNR>96_test[1]..<SNR>96_test2[1]..F[3]"
	v := &Vim{c: cli}
	got, err := v.selectError(msghist)
	if err != nil {
		t.Fatal(err)
	}
	if got.Throwpoint != wantThrowpoint {
		t.Errorf("Vim.selectError(...).Throwpoint = %v, want %v", got.Throwpoint, wantThrowpoint)
	}
}
