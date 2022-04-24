package ttlset

import "time"
import "github.com/spf13/viper"

var defaultTtl time.Duration
var treeOrder int

func init() {
  viper.SetDefault("default_ttl", time.Minute)
  viper.SetDefault("tree_order", 3)

  viper.SetEnvPrefix("ttlset")
  viper.AutomaticEnv()

  defaultTtl = viper.GetDuration("default_ttl")
  treeOrder = viper.GetInt("tree_order")
}
