package main

import (
	"Mango/management/revision"
	"flag"
	"fmt"
)

func main() {
	change := flag.Bool("change", false, "change statu crawling to queued")
	update := flag.Bool("update", false, "change statu finished to queued")

	flag.Parse()

	if *change {
		fmt.Println((*change))
		go revision.Run_statu_revision()
	}
	if *update {
		go revision.Update_statu()
	}
	select {}

}
