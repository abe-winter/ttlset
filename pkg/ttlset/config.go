package ttlset

import "time"
import "github.com/spf13/viper"

var defaultTtl time.Duration
var defaultOrder int

func init() {
  viper.SetDefault("default_ttl", time.Minute)
  viper.SetDefault("default_order", 3)

  viper.SetEnvPrefix("ttlset")
  viper.AutomaticEnv()

  defaultTtl = viper.GetDuration("default_ttl")
  defaultOrder = viper.GetInt("default_order")
}
