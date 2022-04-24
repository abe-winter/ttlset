package main

import "time"
import "github.com/spf13/viper"

var tickerInterval time.Duration

func init() {
  viper.SetDefault("ticker_interval", time.Second * 10)

  viper.SetEnvPrefix("ttlset")
  viper.AutomaticEnv()

  tickerInterval = viper.GetDuration("ticker_interval")
}
