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

package src

import (
	"fmt"
	"github.com/gookit/color"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
)

// CheckRequirements check system requirements to execute this tool
func CheckRequirements() bool {
	// check if qdbus is available
	_, qdbusErr := exec.LookPath("qdbus")
	if qdbusErr != nil {
		color.Errorf("qdbus command is missing - probably it is not installed.\n")
		return false
	}
	_, yakuakeErr := exec.LookPath("yakuake")
	if yakuakeErr != nil {
		color.Errorf("yakuake command is missing - probably it is not installed.\n")
		return false
	}
	// ping yakuake
	executeCmdVoid(DbusPathSessions, DbusMethodPing)
	return true
}

// ReadConfig read configuration and yaml stuff
func ReadConfig(filename string) (*YakCtlConfiguration, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	c := &YakCtlConfiguration{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("YAML syntax error in file: %q: %v", filename, err)
	}
	return c, nil
}

// PrintProfileList prints defined profiles to stdout
func PrintProfileList(configuration *YakCtlConfiguration) {
	if len(*configuration.Profiles) > 0 {
		color.Info.Printf("#\t\tName\n")
	} else {
		color.Warn.Printf("No profiles defined!\n")
	}
	for i, value := range *configuration.Profiles {
		color.Info.Printf("%d\t\t%s\n", i+1, value.Name)
	}
}

// PrintProfile print a single profile in yaml format
func PrintProfile(configuration *YakCtlConfiguration, number int64) error {
	profile, err := GetProfile(configuration, number)
	if err != nil {
		return err
	}
	marshal, ymlErr := yaml.Marshal(profile)
	if ymlErr != nil {
		return fmt.Errorf("problem unmarshalling profile #%d to yaml format")
	}
	color.Info.Println(string(marshal))
	return nil
}

// GetProfile retrieve session struct based on profile number
func GetProfile(configuration *YakCtlConfiguration, number int64) (*ProfileDescription, error) {
	err := fmt.Errorf("profile #%d does not exist", number)
	if number <= 0 || int(number-1) > len(*configuration.Profiles) || len(*configuration.Profiles) == 0 {
		return nil, err
	}
	for i, v := range *configuration.Profiles {
		if int64(i+1) == number {
			return &v, nil
		}
	}
	return nil, err
}

// YakCtlConfiguration this is the configuration object, yaml representation as struct
type YakCtlConfiguration struct {
	Profiles *[]ProfileDescription `yaml:"profiles"`
}

// ProfileDescription represents a session description
type ProfileDescription struct {
	Name       string           `yaml:"name"`
	Tabs       []TabDescription `yaml:"tabs"`
	ClearAll   bool             `yaml:"clear,omitempty"`
	ForceClear bool             `yaml:"force,omitempty"`
}

// TabDescription represents a tab of a yakuake session
type TabDescription struct {
	Name string `yaml:"name"`
	// optional
	Commands  []string `yaml:"commands,omitempty"`
	SplitMode string   `yaml:"split,omitempty"`
	Terminal1 []string `yaml:"terminal1,omitempty"`
	Terminal2 []string `yaml:"terminal2,omitempty"`
	Terminal3 []string `yaml:"terminal3,omitempty"`
	Terminal4 []string `yaml:"terminal4,omitempty"`
	// flags
	Protected            bool `yaml:"protected,omitempty"`
	MonitorSilence       bool `yaml:"monitorSilence,omitempty"`
	MonitorActivity      bool `yaml:"monitorActivity,omitempty"`
	DisableKeyboardInput bool `yaml:"disableInput,omitempty"`
}
