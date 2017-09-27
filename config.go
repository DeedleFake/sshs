package main

import (
	"os"

	"github.com/DeedleFake/wdte"
	"github.com/DeedleFake/wdte/std"
)

func configImporter(from string) (*wdte.Module, error) {
	panic("Not implemented.")
}

func fromConfig(config string) (func(), error) {
	file, err := os.Open(config)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	m, err := new(wdte.Module).Insert(std.Module()).Parse(file, wdte.ImportFunc(configImporter))
	if err != nil {
		return nil, err
	}
}
