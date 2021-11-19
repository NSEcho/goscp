package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pin/tftp"
)

func readHandler(filename string, rf io.ReaderFrom) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := rf.ReadFrom(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	fmt.Printf("%d bytes sent\n", n)
	return nil
}

func writeHandler(filename string, wt io.WriterTo) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := wt.WriteTo(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	fmt.Printf("%d bytes received\n", n)
	return nil
}

func main() {
	s := tftp.NewServer(readHandler, writeHandler)
	s.SetTimeout(5 * time.Second)
	err := s.ListenAndServe(":69")
	if err != nil {
		fmt.Fprintf(os.Stdout, "server: %v\n", err)
		os.Exit(1)
	}
}
