package config

import (
	"flag"
)

var Options struct {
	Login    string
	Password string
}

func ParseOptions() {
	flag.StringVar(&Options.Login, "login", "", "esl  request login")
	flag.StringVar(&Options.Password, "password", "", "password")
	flag.Parse()
}
