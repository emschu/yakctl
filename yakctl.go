/*
 * yakctl - control the yakuake terminal
 *
 * 2020  emschu https://github.com/emschu/yakctl
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/urfave/cli/v2"
	"os"
	"path"
	"strconv"
	"strings"
)

func main() {
	var configFile = ".yakctl.yml"
	var configFilePath string
	var configuration *YakCtlConfiguration
	var verbose bool
	var forceDeletion bool

	hd, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		color.Errorf("%v", homeDirErr)
		return
	}
	configFilePath = path.Join(hd, configFile)

	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "yakctl",
		Usage:                "Control the yakuake terminal and your terminal sessions",
		Version:              "1.0.0",
		HideHelp:             false,
		HideVersion:          false,
		Copyright: "yakctl\t2020\thttps://github.com/emschu/yakctl\n\n" +
			"   This program comes with ABSOLUTELY NO WARRANTY.\n" +
			"   This is free software, and you are welcome\n" +
			"   to redistribute it under the conditions of the\n" +
			"   GPLv3 (https://www.gnu.org/licenses/gpl-3.0.txt).",
		Before: func(context *cli.Context) error {
			// general startup logic
			if verbose {
				fmt.Printf("Using configuration file at: '%s'\n", configFilePath)
			}
			configuration = initApplication(&configFilePath)

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "configuration file",
				Value:       path.Join(hd, ".yakctl.yml"),
				Destination: &configFilePath,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "verbose log output",
				Value:       false,
				Destination: &verbose,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "clear",
				Aliases: []string{"c"},
				Usage:   "Clear all sessions and terminals",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "force",
						Usage:       "force deletion of ALL tabs",
						Value:       false,
						Destination: &forceDeletion,
					},
				},
				Action: func(context *cli.Context) error {
					ClearSession(forceDeletion)
					return nil
				},
			},
			{
				Name:    "profile",
				Aliases: []string{"p"},
				Usage:   "Manage defined profiles, default: list available profiles",
				Action: func(context *cli.Context) error {
					PrintProfileList(configuration)
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:      "show",
						Aliases:   []string{"s"},
						Usage:     "Shows all details of a defined profile",
						ArgsUsage: "profile_id",
						Action: func(context *cli.Context) error {
							profileID, done, _ := getProfileID(context)
							if done {
								return fmt.Errorf("empty profile_id")
							}
							profilePrintErr := PrintProfile(configuration, profileID)
							if profilePrintErr != nil {
								color.Errorf("%v\n", profilePrintErr)
							}
							return nil
						},
					},
					{
						Name:      "open",
						Aliases:   []string{"o"},
						Usage:     "Opens a defined profile",
						ArgsUsage: "profile_id",
						Action: func(context *cli.Context) error {
							profileID, done, err := getProfileID(context)
							if done {
								return err
							}
							if verbose {
								profilePrintErr := PrintProfile(configuration, profileID)
								if profilePrintErr != nil {
									fmt.Printf("%v\n", profilePrintErr)
								}
							}
							err2 := LoadSession(configuration, profileID)
							if err2 != nil {
								return err
							}
							return nil
						},
					},
				},
			},
			{
				Name:      "exec",
				Aliases:   []string{"e"},
				Usage:     "Execute a command in all or specific terminals",
				ArgsUsage: "command to be executed in all or specified terminals",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "terminal",
						Aliases: []string{"t"},
						Usage:   "list Ids of terminals separated by comma and without space",
					},
				},
				Action: func(context *cli.Context) error {
					command := strings.Join(context.Args().Slice(), " ")
					if len(command) == 0 {
						color.Error.Printf("Invalid empty command input detected\n")
						return nil
					}
					color.Info.Printf("Execute '%s' in all terminals\n", command)
					terminalIDInput := strings.Trim(context.String("terminal"), " ")
					var terminalIDs []string
					if len(terminalIDInput) > 0 {
						terminalIDs = strings.Split(terminalIDInput, ",")
					}
					if len(terminalIDs) > 0 {
						color.Info.Printf("Affected terminal ids: %v\n", strings.Join(terminalIDs, ","))
					}
					ExecuteCommand(command, &terminalIDs)
					return nil
				},
			},
			{
				Name:    "status",
				Aliases: []string{"s"},
				Usage:   "List status (=sessions, terminals) of the current yakuake instance",
				Action: func(context *cli.Context) error {
					ShowStatus()
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		color.Errorf("%v\n", err)
	}
}

// get profile_id out of cli arguments
func getProfileID(context *cli.Context) (int64, bool, error) {
	if 0 == context.Args().Len() {
		return 0, true, fmt.Errorf("missing argument 'profile_id'")
	}
	profileID, err := strconv.ParseInt(context.Args().First(), 10, 32)
	if err != nil {
		color.Errorf("Invalid argument 'profile_id'\n")
		return 0, true, err
	}
	return profileID, false, nil
}

// method to handle startup of the application
func initApplication(configFile *string) *YakCtlConfiguration {
	isValid := CheckRequirements()
	if !isValid {
		color.Errorf("Problems detected. yakctl is unable to start.\n")
		os.Exit(1)
	}
	c, confErr := ReadConfig(*configFile)
	if confErr != nil {
		color.Errorf("Invalid configuration detected in configuration file '%s'\n%v\n", *configFile, confErr)
		os.Exit(1)
	}
	return c
}
