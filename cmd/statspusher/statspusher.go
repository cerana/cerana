package main

import "github.com/cerana/cerana/acomm"

type statsPusher struct {
	config  *config
	tracker *acomm.Tracker
}
