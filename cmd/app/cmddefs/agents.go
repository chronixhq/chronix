package cmddefs

import (
	"chronix/cmd/app/rpc"
	serverruntime "chronix/internal/serverruntime"
	"chronix/internal/updater"
	"fmt"
)

type (
	AgentsCommandDef struct {
		Agents AgentsCommand `cmd:"" help:"Manage agents"`
	}

	AgentsCommand struct {
		List     AgentsListCommand     `cmd:"" help:"List connected agents"`
		Update   AgentsUpdateCommand   `cmd:"" help:"Update an agent"`
		Versions AgentsVersionsCommand `cmd:"" help:"List available agent versions" hidden:""`
		Revert   AgentsRevertCommand   `cmd:"" help:"Revert an agent to a specific version" hidden:""`
	}

	AgentsListCommand   struct{}
	AgentsUpdateCommand struct {
		UUID    string `arg:"" help:"Agent UUID or 'all'"`
		Version string `arg:"" optional:"" help:"Target version (defaults to latest)"`
	}
	AgentsVersionsCommand struct{}
	AgentsRevertCommand   struct {
		UUID    string `arg:"" help:"Agent UUID"`
		Version string `arg:"" help:"Version to revert to"`
	}
)

func (c *AgentsListCommand) Run() error {
	var agents []serverruntime.AgentRPCInfo
	if err := rpc.Call("Server.ListAgents", nil, &agents); err != nil {
		return fmt.Errorf("rpc call ListAgents: %w", err)
	}

	fmt.Printf("%-36s %-20s %-10s %-10s %-10s\n", "UUID", "NAME", "VERSION", "STATUS", "UPDATE")
	for _, a := range agents {
		status := "Offline"
		if a.Online {
			status = "Online"
		}
		update := "-"
		if a.UpdateAvailable != "" {
			update = a.UpdateAvailable
		}
		fmt.Printf("%-36s %-20s %-10s %-10s %-10s\n", a.UUID, a.Name, a.Version, status, update)
	}

	return nil
}

func (c *AgentsUpdateCommand) Run() error {
	args := struct {
		UUID    string
		Version string
	}{
		UUID:    c.UUID,
		Version: c.Version,
	}

	if err := rpc.Call("Server.UpdateAgent", &args, nil); err != nil {
		return fmt.Errorf("rpc call UpdateAgent: %w", err)
	}

	if c.UUID == "all" {
		fmt.Println("Update command sent to all online agents.")
	} else {
		fmt.Printf("Update command sent to agent %s.\n", c.UUID)
	}

	return nil
}

func (c *AgentsVersionsCommand) Run() error {
	versions, err := updater.FetchVersions()
	if err != nil {
		return fmt.Errorf("fetch versions: %w", err)
	}

	fmt.Println("Available Agent Versions:")
	for _, ver := range versions.Agent {
		fmt.Printf("- %s (%s)\n", ver.Version, ver.ReleaseDate)
	}

	return nil
}

func (c *AgentsRevertCommand) Run() error {
	args := struct {
		UUID    string
		Version string
	}{
		UUID:    c.UUID,
		Version: c.Version,
	}

	if err := rpc.Call("Server.UpdateAgent", &args, nil); err != nil {
		return fmt.Errorf("rpc call UpdateAgent: %w", err)
	}

	fmt.Printf("Revert command sent to agent %s for version %s.\n", c.UUID, c.Version)
	return nil
}
