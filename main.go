package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	var (
		msg     string
		x, y, z bool
		// addArg  []string
	)

	a := New(os.Args[1:])

	a.BoolVar(&x, "-x", "--xxx")
	a.BoolVar(&y, "-y", "--yyy")
	a.BoolVar(&z, "-z", "--zzz")
	a.StringVar(&msg, "-m", "--msg")
	dir := a.String("-d", "--dir")

	non := a.Parse()

	fmt.Println("x", x)
	fmt.Println("y", y)
	fmt.Println("z", z)

	fmt.Println("msg", msg)
	fmt.Println("dir", *dir)

	fmt.Println(non)
}

type option struct {
	optionShort, optionLong string
	pointTo                 interface{}
}

type args struct {
	inputArguments []string
	options        []option
}

func New(inputArguments []string) args {
	if len(inputArguments) < 1 {
		return args{}
	}
	a := args{inputArguments: inputArguments}
	return a
}

var (
	regexOptionShort = regexp.MustCompile(`^-[\w]$`)
	regexOptionLong  = regexp.MustCompile(`^-(-\w+)+$`)
)

func regexParseOptions(option string) error {
	if regexOptionShort.MatchString(option) || regexOptionLong.MatchString(option) {
		return nil
	}
	return errors.New(`Error parsing option: ` + option)
}

func (a *args) findOptionsIndex(optionShort, optionLong string) (int, string) {
	isShortEmpty := optionShort == ""
	isLongEmpty := optionLong == ""
	for i := 0; i < len(a.options); i++ {
		if !isShortEmpty && a.options[i].optionShort == optionShort {
			return i, optionShort
		} else if !isLongEmpty && a.options[i].optionLong == optionLong {
			return i, optionLong
		}
	}
	return -1, ""
}

func (a *args) newOptionsCheck(optionShort, optionLong string) {
	var optionsBit int
	if optionShort != "" {
		optionsBit |= 1
		if regexParseOptions(optionShort) != nil {
			panic(`Error parsing option: ` + optionShort)
		}
	}
	if optionLong != "" {
		optionsBit |= 2
		if regexParseOptions(optionLong) != nil {
			panic(`Error parsing option: ` + optionLong)
		}
	}
	if optionsBit == 0 {
		panic("No option set")
	}
	if _, duplicateOption := a.findOptionsIndex(optionShort, optionLong); duplicateOption != "" {
		panic("Option already set: \"" + duplicateOption + `"`)
	}
}

func (a *args) StringVar(p *string, optionShort, optionLong string) {
	a.newOptionsCheck(optionShort, optionLong)
	newOption := option{
		optionShort: strings.TrimLeft(optionShort, "-"),
		optionLong:  strings.TrimLeft(optionLong, "-"),
		pointTo:     p,
	}
	a.options = append(a.options, newOption)
}

func (a *args) String(optionShort, optionLong string) *string {
	var s string
	a.StringVar(&s, optionShort, optionLong)
	return &s
}

func (a *args) BoolVar(p *bool, optionShort, optionLong string) {
	a.newOptionsCheck(optionShort, optionLong)
	newOption := option{
		optionShort: strings.TrimLeft(optionShort, "-"),
		optionLong:  strings.TrimLeft(optionLong, "-"),
		pointTo:     p,
	}
	a.options = append(a.options, newOption)
}

func (a *args) Bool(optionShort, optionLong string) *bool {
	var s bool
	a.BoolVar(&s, optionShort, optionLong)
	return &s
}

func (a *args) handleArg(optionIndex int, argumentOption string, argumentArgument *string) {
	switch a.options[optionIndex].pointTo.(type) {
	case *bool:
		if argumentArgument != nil {
			fmt.Println("Option", strconv.Quote(argumentOption), "doesn't allow an argument")
			os.Exit(1)
		}
		*a.options[optionIndex].pointTo.(*bool) = true
	case *string:
		if argumentArgument == nil {
			fmt.Println("Option", strconv.Quote(argumentOption), "requires an argument")
			os.Exit(1)
		}
		*a.options[optionIndex].pointTo.(*string) = *argumentArgument
	}
}

func (a *args) isOptBool(i int) bool {
	if _, ok := a.options[i].pointTo.(*bool); ok {
		return true
	}
	return false
}

func (a *args) Parse() []string {
	var nonOptionArguments []string
	for i := 0; i < len(a.inputArguments); i++ {
		switch {
		case strings.HasPrefix(a.inputArguments[i], "--"):
			if a.inputArguments[i] == "--" {
				nonOptionArguments = append(nonOptionArguments, a.inputArguments[i+1:]...)
				i = len(a.inputArguments)
				break
			}

			inputArgumentSplit := strings.SplitN(a.inputArguments[i], "=", 2)
			inputOption := inputArgumentSplit[0]
			optionIndex, _ := a.findOptionsIndex("", inputOption[2:])
			if optionIndex == -1 {
				fmt.Println("Unrecognized option: " + strconv.Quote(inputOption))
				os.Exit(1)
			}

			var inputArgument *string = nil
			if len(inputArgumentSplit) == 2 {
				inputArgument = &inputArgumentSplit[1]
			} else if !a.isOptBool(optionIndex) && len(a.inputArguments[i:]) > 1 {
				i++
				inputArgument = &a.inputArguments[i]
			}
			a.handleArg(optionIndex, inputOption, inputArgument)
		case strings.HasPrefix(a.inputArguments[i], "-"):
			for j := 1; j < len(a.inputArguments[i]); j++ {
				inputOption := string(a.inputArguments[i][j])
				optionIndex, _ := a.findOptionsIndex(inputOption, "")
				if optionIndex == -1 {
					fmt.Println("Unrecognized option: " + strconv.Quote(inputOption))
					os.Exit(1)
				}
				var inputArgument *string = nil
				if !a.isOptBool(optionIndex) {
					if len(a.inputArguments[i][j:]) > 1 {
						s := a.inputArguments[i][j+1:]
						inputArgument = &s
					} else if len(a.inputArguments[i:]) > 1 {
						i++
						inputArgument = &a.inputArguments[i]
					}
					a.handleArg(optionIndex, inputOption, inputArgument)
					break
				}
				a.handleArg(optionIndex, inputOption, inputArgument)
			}
		default:
			nonOptionArguments = append(nonOptionArguments, a.inputArguments[i])
		}
	}

	return nonOptionArguments
}
