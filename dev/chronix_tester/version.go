package main

import "fmt"

var (
	Version      = "dev"
	ReleaseNotes = ""
)

func printVersion() {
	fmt.Println(Version)
	if ReleaseNotes != "" {
		fmt.Printf("\nRelease Notes:\n%s\n", ReleaseNotes)
	}
}
