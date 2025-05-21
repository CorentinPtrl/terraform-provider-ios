// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ntc

import (
	"embed"
	"github.com/sirikothe/gotextfsm"
	"strings"
)

//go:embed ntc-templates/ntc_templates/templates/*
var content embed.FS

func GetTemplate(name string) (string, error) {
	data, err := content.ReadFile("ntc-templates/ntc_templates/templates/" + name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func GetTemplateNames() ([]string, error) {
	dir, err := content.ReadDir("ntc-templates/ntc_templates/templates")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range dir {
		if !strings.HasPrefix(entry.Name(), "cisco_ios") {
			continue
		}
		names = append(names, entry.Name())
	}
	return names, nil
}

func GetTextFSM(name string) (gotextfsm.TextFSM, error) {
	data, err := GetTemplate(name)
	if err != nil {
		return gotextfsm.TextFSM{}, err
	}
	fsm := gotextfsm.TextFSM{}
	err = fsm.ParseString(data)
	return fsm, err
}
