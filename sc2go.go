package main

import (
	"./mpq"
	"fmt"
	"os"
	)

func main() {
	f, err := mpq.Open("/home/tristan/projects/sc2replayanalysis/replays/etoh_vs_derkommissar_sc2analysis.com_1309372855s1173.sc2replay")
	if f == nil {
		fmt.Printf("can't open file; err=%s\n", err.String())
		os.Exit(1)
	}
}