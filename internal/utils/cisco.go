// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package utils

import (
	"github.com/CorentinPtrl/cgnet"
	"strings"
)

func ConfigDevice(marshal string, device *cgnet.Device) error {
	lines := strings.Split(marshal, "\n")
	configs := []string{}
	configs = append(configs, lines...)
	err := device.Configure(configs)
	if err != nil {
		return err
	}
	return err
}
