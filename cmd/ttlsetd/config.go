package main

import "strings"
import "time"
import "github.com/spf13/viper"

type Role int
const (
  ReadWrite Role = iota
  ReadOnly
)

func ParseRole(s string) Role {
  switch s {
  case "rw": return ReadWrite
  case "ro": return ReadOnly
  default: panic("unk role")
  }
}

type Account struct {
  Key string
  Role Role
}

var tickerInterval time.Duration
var accounts []Account

func init() {
  viper.SetDefault("ticker_interval", time.Second * 10)

  viper.SetEnvPrefix("ttlset")
  viper.AutomaticEnv()

  tickerInterval = viper.GetDuration("ticker_interval")
  rawAccounts := viper.GetStringSlice("accounts")
  for _, raw := range rawAccounts {
    // todo: 1.18 use Cut
    tokens := strings.Split(raw, ":")
    accounts = append(accounts, Account{Key: tokens[0], Role: ParseRole(tokens[1])})
  }
}
