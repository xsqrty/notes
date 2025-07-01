package help

import (
	"fmt"
	"github.com/spf13/pflag"
	"reflect"
	"strings"
)

func GetHelp(base interface{}) string {
	args := ""
	envs := getStructEnvs(base)
	result := ""

	pflag.VisitAll(func(flag *pflag.Flag) {
		sh := ""
		def := ""
		usage := ""

		if flag.Shorthand != "" {
			sh = fmt.Sprintf(", -%s", flag.Shorthand)
		}
		if flag.DefValue != "" {
			def = fmt.Sprintf(" (default: %s)", flag.DefValue)
		}
		if flag.Usage != "" {
			usage = fmt.Sprintf("      %s\n", flag.Usage)
		}

		args += fmt.Sprintf("  --%s%s%s\n%s", flag.Name, sh, def, usage)
	})

	if args != "" {
		result += "Command line arguments:\n" + args
	}

	if envs != "" {
		result += "Environment variables:\n" + envs
	}

	return strings.TrimRight(result, "\n")
}

func getStructEnvs(base interface{}) string {
	envs := ""
	confType := reflect.TypeOf(base)

	if confType.Kind() == reflect.Ptr {
		confType = confType.Elem()
	}

	for i := 0; i < confType.NumField(); i++ {
		field := confType.Field(i)
		if field.Type.Kind() == reflect.Struct {
			envs += getStructEnvs(reflect.New(field.Type).Interface())
		}

		env := field.Tag.Get("env")
		if env == "" {
			continue
		}

		envDefault := field.Tag.Get("envDefault")
		envDescription := field.Tag.Get("envDescription")
		def := ""
		usage := ""

		if envDefault != "" {
			def = fmt.Sprintf(" (default: %s)", envDefault)
		}

		if envDescription != "" {
			usage = fmt.Sprintf("      %s\n", envDescription)
		}

		envs += fmt.Sprintf("  %s%s\n%s", env, def, usage)
	}
	return envs
}
