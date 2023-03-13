package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	if len(os.Args) != 2 {
		panic("not enough pargs")
	}

	if err := createCgroup(os.Args[1]); err != nil {
		panic(err)
	}

	fmt.Println("Starting process")
	if err := startProc(os.Args[1], "sleep 120"); err != nil {
		fmt.Println("Found error")
		panic(err)
	}
}

const cgroupRoot = "/sys/fs/cgroup/"

func createCgroup(name string) error {
	if err := os.Mkdir(filepath.Join(cgroupRoot, name), 0755); err != nil {
		return err
	}

	return nil
}

func startProc(cgroup, cmd string) error {
	cmdParts := strings.Split(cmd, " ")

	c := exec.Command(cmdParts[0], cmdParts[1:]...)

	if err := c.Start(); err != nil {
		return err
	}

	log.Println("PID:", c.Process.Pid)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := c.Wait(); err != nil {
			log.Println(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		f, err := os.OpenFile(filepath.Join(cgroupRoot, cgroup, "cgroup.procs"), os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Println(err)
			return
		}

		if _, err := f.WriteString(fmt.Sprintf("%d", os.Getpid())); err != nil {
			log.Println(err)
			return
		}

		if _, err := f.WriteString(fmt.Sprintf("%d", c.Process.Pid)); err != nil {
			log.Println(err)
			return
		}

		if err := f.Close(); err != nil {
			log.Println(err)
			return
		}
	}()

	wg.Wait()

	return nil
}
