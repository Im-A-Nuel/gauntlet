package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Im-A-Nuel/gauntlet/internal/bob"
	"github.com/Im-A-Nuel/gauntlet/internal/config"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Write .gauntlet/config.yaml, install the Bob Stop hook, and the strengthen-tests Skill",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := rootRepo()
			if err != nil {
				die(ExitConfigOrEnvError, "%v", err)
			}

			if config.Exists(repoRoot) {
				fmt.Printf("%s already exists, leaving it in place\n", config.Path)
			} else {
				if err := config.Write(repoRoot, config.Default()); err != nil {
					die(ExitConfigOrEnvError, "write config: %v", err)
				}
				fmt.Printf("wrote %s\n", config.Path)
			}

			hookChanged, err := bob.InstallHook(repoRoot)
			if err != nil {
				die(ExitConfigOrEnvError, "install Bob hook: %v", err)
			}
			if hookChanged {
				fmt.Printf("installed Bob Stop hook (%s) into %s\n", bob.HookCommand, bob.SettingsPath)
			} else {
				fmt.Printf("Bob Stop hook already present in %s\n", bob.SettingsPath)
			}

			if err := bob.InstallSkill(repoRoot); err != nil {
				die(ExitConfigOrEnvError, "install strengthen-tests skill: %v", err)
			}
			fmt.Printf("wrote %s\n", bob.SkillPath)

			fmt.Println("\nNote: the Bob Stop hook and headless `bob run` invocation are implemented against")
			fmt.Println("IBM's public docs but have not been exercised against a real Bob install in this")
			fmt.Println("environment (docs/CLAUDE_PROGRESS.md has the exact verification status).")
			return nil
		},
	}
}
