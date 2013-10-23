package main

import (
	"Mango/management/jobs"
	"flag"
	"fmt"
)

func main() {
	change := flag.Bool("change", false, "change statu crawling to queued")
	update := flag.Bool("update", false, "change statu finished to queued")

	flag.Parse()

	if *change {
		fmt.Println((*change))
		go jobs.Run_statu_revision()
	}
	if *update {
		go jobs.Update_statu()
	}
	select {}

}
