package svc

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"
)

type Config struct {
	Name        string
	DisplayName string
	Description string
	Arguments   []string
}

var (
	// ServiceRunning is true if the application is running as a service.
	ServiceRunning bool
	// Svc is the service instance.
	Svc service.Service
	// OnStop is called when the service is requested to stop.
	OnStop func()
)

func Handle(cfg Config, action string) error {
	svcConfig := &service.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
		Arguments:   cfg.Arguments,
		Option: map[string]interface{}{
			"StartType": "automatic",
		},
	}

	prg := &program{}
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
		case service.StatusUnknown:
			statusStr = "unknown"
		}
		fmt.Printf("Service %s is %s.\n", cfg.Name, statusStr)
		return nil
	default:
		return fmt.Errorf("invalid action: %s", action)
	}
}

func RunIfService(name string) {
	if service.Interactive() {
		return
	}

	svcConfig := &service.Config{
		Name: name,
	}
	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return
	}
	Svc = s
	ServiceRunning = true
	go func() {
		_ = s.Run()
	}()
	// Wait a bit to ensure s.Run() has started the dispatcher
	time.Sleep(100 * time.Millisecond)
}

type program struct{}

func (p *program) Start(_ service.Service) error {
	return nil
}

func (p *program) Stop(_ service.Service) error {
	if OnStop != nil {
		OnStop()
	}
	return nil
}
