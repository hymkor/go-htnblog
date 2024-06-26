package flag

// flexible-flag package
//
// On the standard version of "flag", parsing stops just before the first
// non-flag argument.
//
// On this version, it continues for compatiblity for v0.9.0

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var Debug io.Writer = io.Discard

//var Debug io.Writer = os.Stderr

type _StringFlag struct {
	_name    string
	_default string
	_usage   string
	_value   string
}

var ErrTooFewArgument = errors.New("too few arguments")

func (s *_StringFlag) parse(args []string, log io.Writer) ([]string, error) {
	if len(args) <= 0 {
		return nil, ErrTooFewArgument
	}
	s._value = args[0]
	fmt.Fprintf(Debug, "%s: set %#v\n", s._name, s._value)
	return args[1:], nil
}

func (s *_StringFlag) usage() string {
	var u strings.Builder
	fmt.Fprintf(&u, "  %s string", s._name)
	if u.Len() <= 6 {
		u.WriteByte('\t')
	} else {
		u.WriteString("\n    \t")
	}
	u.WriteString(s._usage)
	return u.String()
}

type _BoolFlag struct {
	_name    string
	_default bool
	_usage   string
	_value   bool
}

func (b *_BoolFlag) parse(args []string, log io.Writer) ([]string, error) {
	b._value = true
	fmt.Fprintf(Debug, "%s: set %#v\n", b._name, b._value)
	return args, nil
}

func (b *_BoolFlag) usage() string {
	var u strings.Builder
	fmt.Fprintf(&u, "  %s", b._name)
	if u.Len() <= 6 {
		u.WriteByte('\t')
	} else {
		u.WriteString("\n    \t")
	}
	u.WriteString(b._usage)
	return u.String()
}

type _IntFlag struct {
	_name    string
	_default int
	_usage   string
	_value   int
}

func (i *_IntFlag) parse(args []string, log io.Writer) ([]string, error) {
	if len(args) <= 0 {
		return nil, ErrTooFewArgument
	}
	var err error
	i._value, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(Debug, "%s: set %#v\n", i._name, i._value)
	return args[1:], nil
}

func (i *_IntFlag) usage() string {
	var u strings.Builder
	fmt.Fprintf(&u, "  %s int", i._name)
	if u.Len() <= 6 {
		u.WriteByte('\t')
	} else {
		u.WriteString("\n    \t")
	}
	u.WriteString(i._usage)
	return u.String()
}

type _Flag interface {
	parse([]string, io.Writer) ([]string, error)
	usage() string
}

type FlagSet struct {
	flags      map[string]_Flag
	nonOptions []string
}

func (f *FlagSet) String(name, defaults, usage string) *string {
	o := &_StringFlag{
		_name:    "-" + name,
		_default: defaults,
		_usage:   usage,
		_value:   defaults,
	}
	if f.flags == nil {
		f.flags = make(map[string]_Flag)
	}
	f.flags[name] = o
	return &o._value
}

func (f *FlagSet) Bool(name string, defaults bool, usage string) *bool {
	b := &_BoolFlag{
		_name:    "-" + name,
		_default: defaults,
		_usage:   usage,
		_value:   defaults,
	}
	if f.flags == nil {
		f.flags = make(map[string]_Flag)
	}

	f.flags[name] = b
	return &b._value
}

func (f *FlagSet) Int(name string, defaults int, usage string) *int {
	i := &_IntFlag{
		_name:    "-" + name,
		_default: defaults,
		_usage:   usage,
		_value:   defaults,
	}
	if f.flags == nil {
		f.flags = make(map[string]_Flag)
	}
	f.flags[name] = i
	return &i._value
}

func (f *FlagSet) Args() []string {
	return f.nonOptions
}

func (f *FlagSet) PrintDefaults() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	for _, value := range f.flags {
		fmt.Fprintln(os.Stderr, value.usage())
	}
}

func (f *FlagSet) Parse(args []string) error {
	if f.flags == nil {
		f.flags = map[string]_Flag{}
	}
	for len(args) > 0 {
		name := args[0]
		args = args[1:]
		if len(name) <= 0 || name[0] != '-' {
			f.nonOptions = append(f.nonOptions, name)
			continue
		}
		o, ok := f.flags[name[1:]]
		if !ok {
			if name == "-h" {
				f.PrintDefaults()
				os.Exit(0)
			}
			return fmt.Errorf("flag provided but not defined: %s", name)
		}
		var err error
		args, err = o.parse(args, os.Stderr)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	fmt.Fprintf(Debug, "Non Option args: %#v\n", f.nonOptions)
	return nil
}

var CommandLine FlagSet

func String(name, defaults, usage string) *string {
	return CommandLine.String(name, defaults, usage)
}

func Bool(name string, defaults bool, usage string) *bool {
	return CommandLine.Bool(name, defaults, usage)
}

func Int(name string, defaults int, usage string) *int {
	return CommandLine.Int(name, defaults, usage)
}

func Args() []string {
	return CommandLine.Args()
}

func PrintDefaults() {
	CommandLine.PrintDefaults()
}

func Parse() {
	err := CommandLine.Parse(os.Args[1:])
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err.Error())
	PrintDefaults()
	os.Exit(1)
}
