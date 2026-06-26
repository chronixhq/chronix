package main

import cfgpkg "chronix-agent/agentconfig"

type agentConfig = cfgpkg.Config

func defaultConfigPath() string {
	return cfgpkg.DefaultPath()
}

func loadConfig(path string) (*agentConfig, error) {
	return cfgpkg.Load(path)
}

func saveConfig(path string, cfg *agentConfig) error {
	return cfgpkg.Save(path, cfg)
}
