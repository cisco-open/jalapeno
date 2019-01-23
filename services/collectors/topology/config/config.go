//Package config imports configuration from the command line, ENV vars and config file
package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/database"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/kafka"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ErrConfigFileNotFound = errors.New("Global Config file not found or specified.")
)

type GlobalConfig struct {
	Log    log.Config
	Arango database.ArangoConfig
}


type TopologyConfig struct {
	Kafka kafka.Config
	ASN   string
        DirectPeerASNS string
        TransitProviderASNS string
}

func InitGlobalCfg() *GlobalConfig {
	return &GlobalConfig{
		Log:    log.NewLogrConfig(),
		Arango: database.NewConfig(),
	}
}

func InitTopologyCfg() *TopologyConfig {
	return &TopologyConfig{
		Kafka: kafka.NewConfig(),
	}
}

// InitFlags: sets up all command flags
func InitGlobalFlags(ccmd *cobra.Command, cfg interface{}) error {
	// Add non-struct flags here
	ccmd.PersistentFlags().String("config", "", "The configuration file")
	viper.BindPFlag("config", ccmd.PersistentFlags().Lookup("config"))

	return setupEnvAndFlags(ccmd, cfg)
}

func InitFlags(ccmd *cobra.Command, cfg interface{}) error {
	return setupEnvAndFlags(ccmd, cfg)
}

func Reset() {
	viper.Reset()
}

// Getconfig loads the config file
func GetConfig(cfg interface{}) (interface{}, error) {

	//override with viper config file
	configFile := viper.GetString("config")
	if configFile != "" {
		if _, err := os.Stat(configFile); err == nil {
			viper.SetConfigFile(configFile) // name of config file
			err := viper.ReadInConfig()     // Find and read the config file
			if err != nil {                 // Handle errors reading the config file
				return cfg, err
			}
			err = viper.Unmarshal(cfg)
			if err != nil { // Handle errors reading the config file
				return cfg, err
			}
		}
	}
	err := getCfg(cfg)
	return cfg, err
}

/* NOTE: Due to a bug in Viper, all boolean flags MUST DEFAULT TO FALSE.
   That is, all boolean flags should be to ENABLE features.
	 --use-db vs --dont-use-db.
*/

func getCfg(cfg interface{}) error {
	return eachSubField(cfg, func(parent reflect.Value, subFieldName string, crumbs []string) error {
		p := strings.Join(crumbs, "")
		envStr := envString(p, subFieldName)
		flagStr := flagString(p, subFieldName)

		// eachSubField only calls this function if  subFieldName exists
		// and can be set
		subField := parent.FieldByName(subFieldName)

		str := ""
		if v := viper.Get(envStr); v != nil {
			str = envStr
		} else if viper.Get(flagStr) != nil {
			str = flagStr
		}
		if len(str) != 0 && subField.CanSet() {
			switch subField.Type().Kind() {
			case reflect.Bool:
				v := viper.GetBool(str)
				subField.SetBool(v || subField.Bool()) // IsSet is broken with bools, see NOTE above ^^^
			case reflect.Int:
				v := viper.GetInt(str)
				if v == 0 {
					return nil
				}
				subField.SetInt(int64(v))
			case reflect.Int64:
				v := viper.GetInt64(str)
				if v == 0 {
					return nil
				}
				subField.SetInt(v)
			case reflect.String:
				v := viper.GetString(str)
				if len(v) == 0 {
					return nil
				}
				subField.SetString(v)
			case reflect.Float64:
				v := viper.GetFloat64(str)
				if v == 0 {
					return nil
				}
				subField.SetFloat(v)
			case reflect.Slice:
				v := viper.GetStringSlice(str)
				if len(v) == 0 || len(v[0]) == 0 {
					return nil
				}
				subField.Set(reflect.ValueOf(v))
			default:
				return fmt.Errorf("%s is unsupported by config @ %s.%s", subField.Type().String(), p, subFieldName)
			}
		}
		return nil
	})
}

// Process env var overrides for all values
func setupEnvAndFlags(voltronCmd *cobra.Command, cfg interface{}) error {
	// Supports fetching value from env for all config of type: int, float64, bool, and string
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("voltron")

	return eachSubField(cfg, func(parent reflect.Value, subFieldName string, crumbs []string) error {
		p := strings.Join(crumbs, "")
		envStr := envString(p, subFieldName)
		flagStr := flagString(p, subFieldName)
		viper.BindEnv(envStr)

		subField, _ := parent.Type().FieldByName(subFieldName)

		desc := subField.Tag.Get("desc")
		switch subField.Type.Kind() {
		case reflect.Bool:
			voltronCmd.PersistentFlags().Bool(flagStr, false, desc)
		case reflect.Int:
			voltronCmd.PersistentFlags().Int(flagStr, 0, desc)
		case reflect.Int64:
			voltronCmd.PersistentFlags().Int64(flagStr, 0, desc)
		case reflect.String:
			voltronCmd.PersistentFlags().String(flagStr, "", desc)
		case reflect.Float64:
			voltronCmd.PersistentFlags().Float64(flagStr, 0, desc)
		case reflect.Slice:
			if subField.Type.Elem().Kind() == reflect.String {
				voltronCmd.PersistentFlags().StringSlice(flagStr, nil, desc)
			} else {
				return fmt.Errorf("%s is unsupported by config @ %s.%s", subField.Type.String(), p, subFieldName)
			}
		default:
			return fmt.Errorf("%s is unsupported by config @ %s.%s", subField.Type.String(), p, subFieldName)
		}
		viper.BindPFlag(flagStr, voltronCmd.PersistentFlags().Lookup(flagStr))
		return nil
	})
}

// eachSubField is used for a struct of structs (like GlobalConfig). fn is called
// with each field from each sub-struct of the parent. Fields are skipped if they
// are not settable, or unexported OR are marked with `flag:"false"`
func eachSubField(i interface{}, fn func(reflect.Value, string, []string) error, crumbs ...string) error {
	t := reflect.ValueOf(i)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		panic("eachSubField can only be called on a pointer-to-struct")
	}
	// Sanity check. Should be true if it is a pointer-to-struct
	if !t.Elem().CanSet() {
		panic("eachSubField can only be called on a settable struct of structs")
	}

	t = t.Elem()
	nf := t.NumField()
	for i := 0; i < nf; i++ {
		field := t.Field(i)
		sf := t.Type().Field(i)
		if sf.Tag.Get("flag") == "false" {
			continue
		}

		if field.Kind() == reflect.Struct && field.CanSet() {
			eachSubField(field.Addr().Interface(), fn, append(crumbs, sf.Name)...)
		} else if field.CanSet() {
			if err := fn(t, sf.Name, crumbs); err != nil {
				return err
			}
		}
	}
	return nil
}

func flagString(parent, field string) string {
	if len(parent) == 0 {
		return strings.ToLower(hyphen(field))
	}
	if len(field) == 0 {
		return strings.ToLower(hyphen(parent))
	}
	return strings.ToLower(hyphen(parent) + "-" + hyphen(field))
}

func envString(parent, field string) string {
	if len(parent) == 0 {
		return strings.ToUpper(underscore(field))
	}
	if len(field) == 0 {
		return strings.ToUpper(underscore(parent))
	}
	return strings.ToUpper(underscore(parent) + "_" + underscore(field))
}

func isUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

func isLower(c byte) bool {
	return !isUpper(c)
}

func toUpper(c byte) byte {
	return c - ('a' - 'A')
}

func toLower(c byte) byte {
	return c + ('a' - 'A')
}

// Underscore converts "CamelCasedString" to "camel_cased_string".
func underscore(s string) string {
	return splitCamel(s, '_')
}

// Underscore converts "CamelCasedString" to "camel-cased-string".
func hyphen(s string) string {
	return splitCamel(s, '-')
}

func splitCamel(s string, sep byte) string {
	r := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isUpper(c) {
			if i > 0 && i+1 < len(s) && (isLower(s[i-1]) || isLower(s[i+1])) {
				r = append(r, sep, toLower(c))
			} else {
				r = append(r, toLower(c))
			}
		} else {
			r = append(r, c)
		}
	}
	return string(r)
}
