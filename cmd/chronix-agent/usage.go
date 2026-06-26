package main

import (
	"fmt"
)

func printTopUsage() {
	cfgPath := defaultConfigPath()
	fmt.Println("Chronix Agent")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chronix-agent [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  register   Register this machine as an agent (one-time setup)")
	fmt.Println("  info       Show the current agent registration details")
	fmt.Println("  repoint    Update the saved server address without changing identity")
	fmt.Println("  repair-register Refresh or recover registration using the current UUID and keys")
	fmt.Println("  rename     Change the agent name and refresh the server record")
	fmt.Println("  run        Start the agent using the saved registration (default)")
	fmt.Println("  stop       Stop the running agent")
	fmt.Println("  status     Check if the agent is running")
	fmt.Println("  restart    Restart the running agent")
	fmt.Println("  reset      Clear the saved server pin")
	fmt.Println("  unregister Remove the local registration and decommission it when possible")
	fmt.Println("  version    Print the version")
	fmt.Println()
	fmt.Println("Service Commands:")
	fmt.Println("  service install      Install as a native system service (macOS/Linux require sudo)")
	fmt.Println("  service uninstall    Uninstall the native service (macOS/Linux require sudo)")
	fmt.Println("  service start        Start the service (macOS/Linux require sudo)")
	fmt.Println("  service stop         Stop the service (macOS/Linux require sudo)")
	fmt.Println("  service status       Check the service status")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  chronix-agent register 192.168.0.102 MyAgent")
	fmt.Println("  chronix-agent service status")
	fmt.Println("  chronix-agent status")
	fmt.Println()
	fmt.Println("Download latest:")
	fmt.Println("  https://dist.chronixhq.com/latest/<platform>/chronix-agent")
	fmt.Println("  Platforms: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Registration contacts the server UI at https://<server>:5170 for approval.")
	fmt.Println("  - Agent connects to the server over WSS on port 5172 by default.")
	fmt.Printf("  - Config path: %s\n", cfgPath)
}

func printServiceUsage() {
	fmt.Println("Manage the native system service")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chronix-agent service [action]")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  install    Install as a native system service (macOS/Linux require sudo)")
	fmt.Println("  uninstall  Uninstall the native service (macOS/Linux require sudo)")
	fmt.Println("  start      Start the service (macOS/Linux require sudo)")
	fmt.Println("  stop       Stop the service (macOS/Linux require sudo)")
	fmt.Println("  status     Check the service status")
}

func printRegisterUsage() {
	fmt.Println("Register the agent with the Chronix server")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chronix-agent register [flags] <server[:port]> <name>")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server  The Chronix server host or host:port (default port 5172)")
	fmt.Println("  -name    A unique display name for this agent")
	fmt.Println("  -force   Create a new UUID and keypair even if already registered")
	fmt.Println()
	fmt.Println("Behavior:")
	fmt.Println("  - Generates an Ed25519 keypair and a UUID for the agent.")
	fmt.Println("  - Requests approval in the Chronix web UI (5-minute window).")
	fmt.Println("  - On approval, saves the registration locally and starts the agent.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  chronix-agent register localhost MyAgent")
	fmt.Println("  chronix-agent register -force myhost:5172 MyAgent")
}

func printResetUsage() {
	fmt.Println("Clear the saved server pin")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chronix-agent reset [-y]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -y   Skip confirmation prompt")
	fmt.Println()
	fmt.Println("Behavior:")
	fmt.Println("  - Prompts for confirmation (type 'RESET') and clears the saved server fingerprint.")
	fmt.Println("  - The next connection will trust the server and pin its fingerprint again (TOFU).")
}

func printUnregisterUsage() {
	fmt.Println("Remove the local agent registration")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chronix-agent unregister [-y]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -y   Skip confirmation prompt")
	fmt.Println()
	fmt.Println("Behavior:")
	fmt.Println("  - Prompts for confirmation (type 'UNREGISTER') and removes the local registration file.")
	fmt.Println("  - Best-effort notifies the current server and stops the local agent first when possible.")
	fmt.Println("  - This deletes the local agent identity (UUID and keys). A future register creates a new identity.")
}

func printInfoUsage() {
	fmt.Println("Show the current agent registration details")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chronix-agent info")
	fmt.Println()
	fmt.Println("Behavior:")
	fmt.Println("  - Prints the saved name, UUID, server target, config file, and server pin status.")
	fmt.Println("  - Reports whether the local agent is running and performs a quick server probe.")
}

func printRepointUsage() {
	fmt.Println("Update the saved server address")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chronix-agent repoint <server[:port]>")
	fmt.Println()
	fmt.Println("Behavior:")
	fmt.Println("  - Updates only the saved server host and port.")
	fmt.Println("  - Keeps the existing UUID, keys, name, and server pin.")
	fmt.Println("  - Restarts the running agent automatically when possible.")
}

func printRepairRegisterUsage() {
	fmt.Println("Refresh or recover registration using the current UUID and keys")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chronix-agent repair-register [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server  Override the saved server host or host:port")
	fmt.Println("  -name    Override the saved agent display name")
	fmt.Println()
	fmt.Println("Behavior:")
	fmt.Println("  - Reuses the existing UUID and keypair.")
	fmt.Println("  - Optionally updates the saved server target or name before refreshing the registration.")
	fmt.Println("  - Tries an authenticated repair first, then falls back to approval-based recovery if needed.")
	fmt.Println("  - Refreshes the saved server pin from the responding server.")
}

func printRenameUsage() {
	fmt.Println("Change the agent name and refresh the server record")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chronix-agent rename <new-name>")
	fmt.Println()
	fmt.Println("Behavior:")
	fmt.Println("  - Updates the local agent name.")
	fmt.Println("  - Refreshes the server-side agent record using the existing UUID and keypair.")
}

func printRunUsage() {
	fmt.Println("Start the agent using the saved registration")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chronix-agent run")
	fmt.Println()
	fmt.Println("Behavior:")
	fmt.Println("  - Connects to the server using the configuration saved during registration.")
	fmt.Println("  - This is the default action when no command is provided.")
}

func printVersionUsage() {
	fmt.Println("Print the agent version")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chronix-agent version")
}
