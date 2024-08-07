package main

import (
	"testing"
)

func TestL4d2(t *testing.T) {

	extractVPK(
		"S:\\Program Files\\Steam\\steamapps\\common\\Left 4 Dead 2\\left4dead2\\pak01_dir.vpk",
		"W:\\GameContent\\l4d2", []string{"*.mdl"}, 0)
}
