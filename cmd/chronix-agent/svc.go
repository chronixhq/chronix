package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"
)

type SvcConfig struct {
	Name        string
	DisplayName string
	Description string
	Arguments   []string
}

var (
	// agentServiceRunning is true if the application is running as a service.
	agentServiceRunning bool
	// agentService is the service instance.
	agentService service.Service
	// onAgentStop is called when the service is requested to stop.
	onAgentStop func()
)

func handleAgentService(cfg SvcConfig, action string) error {
	svcConfig := &service.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
		Arguments:   cfg.Arguments,
		Option: map[string]interface{}{
			"StartType": "automatic",
		},
	}

	prg := &agentProgram{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return err
	}

	switch action {
	case "install":
		err = s.Install()
		if err != nil {
			if runtime.GOOS != "windows" && (os.IsPermission(err) || strings.Contains(err.Error(), "permission denied")) {
				return fmt.Errorf("%w\nTIP: Try running this command with 'sudo'", err)
			}
			return err
		}
		fmt.Printf("Service %s installed successfully.\n", cfg.Name)
		_ = s.Start()
		return nil
	case "uninstall":
		_ = s.Stop()
		err = s.Uninstall()
		if err != nil {
			if runtime.GOOS != "windows" && (os.IsPermission(err) || strings.Contains(err.Error(), "permission denied")) {
				return fmt.Errorf("%w\nTIP: Try running this command with 'sudo'", err)
			}
			return err
		}
		fmt.Printf("Service %s uninstalled successfully.\n", cfg.Name)
		return nil
	case "start":
		err = s.Start()
		if err != nil {
			if runtime.GOOS != "windows" && (os.IsPermission(err) || strings.Contains(err.Error(), "permission denied")) {
				return fmt.Errorf("%w\nTIP: Try running this command with 'sudo'", err)
			}
			return err
		}
		fmt.Printf("Service %s started successfully.\n", cfg.Name)
		return nil
	case "stop":
		err = s.Stop()
		if err != nil {
			if runtime.GOOS != "windows" && (os.IsPermission(err) || strings.Contains(err.Error(), "permission denied")) {
				return fmt.Errorf("%w\nTIP: Try running this command with 'sudo'", err)
			}
			return err
		}
		fmt.Printf("Service %s stopped successfully.\n", cfg.Name)
		return nil
	case "status":
		status, err := s.Status()
		if err != nil {
			return err
		}
		statusStr := "unknown"
		switch status {
		case service.StatusRunning:
			statusStr = "running"
		case service.StatusStopped:
			statusStr = "stopped"
		}
		fmt.Printf("Service %s is %s.\n", cfg.Name, statusStr)
		return nil
	default:
		return fmt.Errorf("invalid action: %s", action)
	}
}

func runAgentIfService(name string) {
	if service.Interactive() {
		return
	}

	svcConfig := &service.Config{
		Name: name,
	}
	prg := &agentProgram{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return
	}
	agentService = s
	agentServiceRunning = true
	go func() {
		_ = s.Run()
	}()
	// Wait a bit to ensure s.Run() has started the dispatcher
	time.Sleep(100 * time.Millisecond)
}

type agentProgram struct{}

func (p *agentProgram) Start(_ service.Service) error {
	return nil
}

func (p *agentProgram) Stop(_ service.Service) error {
	if onAgentStop != nil {
		onAgentStop()
	}
	return nil
}
