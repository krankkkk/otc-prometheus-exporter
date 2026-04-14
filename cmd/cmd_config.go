package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

/*
initializeConfig is a helper function which sets the environment variable for a flag. It gives precedence to the flag,
meaning that the env is only taken if the flag is empty. It assigns the environment variables to the flags which are
defined in the map flagToEnvMap.
*/
func InitializeConfig(cmd *cobra.Command, flagToEnvMapping map[string]string) error {
	v := viper.New()
	v.AutomaticEnv()

	var setErr error
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if setErr != nil {
			return
		}
		configName, ok := flagToEnvMapping[f.Name]
		if !ok {
			return
		}
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				setErr = fmt.Errorf("flag %q from env %q: %w", f.Name, configName, err)
			}
		}
	})
	if setErr != nil {
		return setErr
	}

	return nil
}
