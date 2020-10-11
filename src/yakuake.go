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
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// remember: a dbus cmd consists of service + path + interface method
const (
	DbusApp     = "qdbus"
	DbusService = "org.kde.yakuake"
	// paths
	DbusPathSessions   = "/yakuake/sessions"
	DbusPathTabs       = "/yakuake/tabs"
	DbusPathWindow     = "/yakuake/window"
	DbusPathMainwindow = "/yakuake/MainWindow_1"
	// methods for paths = sessions
	DbusMethodAddSession                 = "org.kde.yakuake.addSession"
	DbusMethodAddSessionLr               = "org.kde.yakuake.addSessionTwoHorizontal"
	DbusMethodAddSessionTb               = "org.kde.yakuake.addSessionTwoVertical"
	DbusMethodAddSessionQu               = "org.kde.yakuake.addSessionQuad"
	DbusMethodSetSessionMonitorSilence   = "org.kde.yakuake.setSessionMonitorSilenceEnabled"
	DbusMethodSetSessionMonitorActivity  = "org.kde.yakuake.setSessionMonitorActivityEnabled"
	DbusMethodSetKeyboardInputEnabled    = "org.kde.yakuake.setSessionKeyboardInputEnabled"
	DbusMethodTerminalIDList             = "org.kde.yakuake.terminalIdList"
	DbusMethodTerminalIDListForSessionID = "org.kde.yakuake.terminalIdsForSessionId"
	DbusMethodTerminalRemoval            = "org.kde.yakuake.removeTerminal"
	DbusMethodSessionIDForTerminalID     = "org.kde.yakuake.sessionIdForTerminalId"
	DbusMethodSessionIDList              = "org.kde.yakuake.sessionIdList"
	DbusMethodIsSessionClosable          = "org.kde.yakuake.isSessionClosable"
	DbusMethodSetSessionClosable         = "org.kde.yakuake.setSessionClosable"
	DbusMethodRunCommandInTerminal       = "org.kde.yakuake.runCommandInTerminal"
	DbusMethodActiveSessionId            = "org.kde.yakuake.activeSessionId"

	// methods for paths = tabs
	DbusMethodTabTitle    = "org.kde.yakuake.tabTitle"
	DbusMethodSetTabTitle = "org.kde.yakuake.setTabTitle"

	// methods for path = window
	DbusMethodToggleState = "org.kde.yakuake.toggleWindowState"

	// methods for path = MainWindow_1
	DbusMethodQwidgetVisible = "org.qtproject.Qt.QWidget.visible"

	DbusMethodPing = "org.freedesktop.DBus.Peer.Ping"
)

// LoadSession method to load a yakuake session defined in yaml configuration
func LoadSession(configuration *YakCtlConfiguration, profileID int64) error {
	profile, err := GetProfile(configuration, profileID)
	if err != nil {
		return err
	}

	currentlyOpenedSessionID := getCurrentSessionId()

	// store these ids for later to avoid killing the shell we possibly run in
	openedTerminalsBeforeLoad, termErrs := getAllTerminalIDs()
	if termErrs != nil {
		return termErrs
	}

	for _, tab := range profile.Tabs {
		sessionID := startSession(&tab)
		if sessionID == nil || len(*sessionID) == 0 {
			return fmt.Errorf("problem creating session for new tab")
		}

		// set title
		executeCmdVoid(DbusPathTabs, DbusMethodSetTabTitle, *sessionID, tab.Name)

		// get terminal ids of session
		terminalIDs := getTerminalIDsForSessionID(sessionID)

		// commands are executed on each terminal, before the specific stuff commands will be executed
		for _, command := range tab.Commands {
			for _, terminalID := range terminalIDs {
				executeCommandInTerminal(command, terminalID)
			}
		}
		// handle different terminals
		if len(terminalIDs) > 0 {
			for _, t1Cmd := range tab.Terminal1 {
				executeCommandInTerminal(t1Cmd, terminalIDs[0])
			}
		}
		if len(terminalIDs) > 1 {
			for _, t2Cmd := range tab.Terminal2 {
				executeCommandInTerminal(t2Cmd, terminalIDs[1])
			}
		}
		if len(terminalIDs) > 2 {
			for _, t3Cmd := range tab.Terminal3 {
				executeCommandInTerminal(t3Cmd, terminalIDs[2])
			}
		}
		if len(terminalIDs) > 3 {
			for _, t4Cmd := range tab.Terminal4 {
				executeCommandInTerminal(t4Cmd, terminalIDs[3])
			}
		}
		// handle flags
		if tab.Protected {
			executeCmdVoid(DbusPathSessions, DbusMethodSetSessionClosable, *sessionID, "false")
		}
		if tab.MonitorSilence {
			executeCmdVoid(DbusPathSessions, DbusMethodSetSessionMonitorSilence, *sessionID, "true")
		}
		if tab.MonitorActivity {
			executeCmdVoid(DbusPathSessions, DbusMethodSetSessionMonitorActivity, *sessionID, "true")
		}
		if tab.DisableKeyboardInput {
			executeCmdVoid(DbusPathSessions, DbusMethodSetKeyboardInputEnabled, *sessionID, "false")
		}
		color.Success.Printf("Created new session #%s\n", *sessionID)
	}

	// toggle window
	if !isWindowShown() {
		// toggle window
		executeCmdVoid(DbusPathWindow, DbusMethodToggleState)
	}

	// clean up
	if profile.ClearAll {
		clearSessions(profile.ForceClear, openedTerminalsBeforeLoad, &currentlyOpenedSessionID)
	}

	return nil
}

// ClearSession method to reset yakuake
func ClearSession(forceDeletion bool) {
	// get all terminal Ids and remove them afterwards
	var err error
	terminalIDs, err := getAllTerminalIDs()
	if err != nil {
		color.Error.Printf("%v.\n", err)
		return
	}

	clearSessions(forceDeletion, terminalIDs, nil)
}

// ExecuteCommand method to execute a command in all or in specified terminals
func ExecuteCommand(command string, affectedTerminals *[]string) {
	if len(*affectedTerminals) == 0 {
		var err error
		*affectedTerminals, err = getAllTerminalIDs()
		if err != nil {
			color.Error.Printf("%v\n", err)
			return
		}
	}
	for _, tID := range *affectedTerminals {
		executeCommandInTerminal(command, tID)
	}
}

// ShowStatus show sessions and terminals of the current yakuake instance
func ShowStatus() {
	sessionIDs, err := getAllSessionIDs()
	if err != nil {
		color.Errorf("%v\n", sessionIDs)
		return
	}
	for _, sessionID := range sessionIDs {
		tabTitle := getTitleOfSession(sessionID)
		color.Info.Printf("- session #%s, tab title: %s\n", sessionID, *tabTitle)
		terminalIds := getTerminalIDsForSessionID(&sessionID)
		for _, terminalID := range terminalIds {
			color.Info.Printf("\t|- Terminal #%s\n", terminalID)
		}
	}
}

// clear the specified terminals
func clearSessions(forceDeletion bool, terminalIDs []string, lastSessionId *string) {
	var currentlyActiveSessionID string
	if lastSessionId == nil {
		currentlyActiveSessionID = getCurrentSessionId()
	} else {
		currentlyActiveSessionID = *lastSessionId
	}

	if len(terminalIDs) == 0 {
		color.Info.Printf("Found NO open terminal\n")
	} else if len(terminalIDs) == 1 {
		color.Info.Printf("Found one open terminal that will be tried to close\n")
	} else {
		color.Info.Printf("Found %d open terminals that will be tried to close\n", len(terminalIDs))
	}
	didSomething := false
	if forceDeletion {
		color.Warn.Printf("Closing of tabs will be forced!\n")
	}

	// strip currently active shell from the terminal id slice
	// split the slice into now and postponed
	var isCurrentTerminalPostponed = false
	var cleanedUpTerminalIDList = terminalIDs
	var postponedTerminalIDs []string

	for i, tID := range terminalIDs {
		sessionIDOfTerminal := getSessionIDForTerminalID(tID)
		if sessionIDOfTerminal == currentlyActiveSessionID {
			cleanedUpTerminalIDList = append(terminalIDs[:i], terminalIDs[i+1:]...)
			isCurrentTerminalPostponed = true
			postponedTerminalIDs = append(postponedTerminalIDs, tID)
		}
	}

	processTerminalRemoval(&forceDeletion, &cleanedUpTerminalIDList, &didSomething)

	// remove the currently opened terminal at the end
	if isCurrentTerminalPostponed {
		processTerminalRemoval(&forceDeletion, &postponedTerminalIDs, &didSomething)
	}

	if didSomething {
		color.Success.Println("All sessions cleared!")
	}
}

func getCurrentSessionId() string {
	currentlyActiveSessionID, activeSessionIDErr := executeCmd(DbusPathSessions, DbusMethodActiveSessionId)
	if activeSessionIDErr != nil {
		color.Errorf("problem fetching current active session id")
	}
	if currentlyActiveSessionID == "-1" {
		currentlyActiveSessionID = ""
	}
	return currentlyActiveSessionID
}

func processTerminalRemoval(forceDeletion *bool, terminalIDs *[]string, didSomething *bool) {
	for _, tID := range *terminalIDs {
		closable, title := isTerminalClosable(tID)
		if !closable && !*forceDeletion {
			color.Warn.Printf("Terminal #%s ('%s') is protected and not closable. Do it manually!\n", tID, *title)
		}
		sessionID := getSessionIDForTerminalID(tID)
		if *forceDeletion {
			// set closable by dbus command
			for _, id := range strings.Split(sessionID, ",") {
				executeCmdVoid(DbusPathSessions, DbusMethodSetSessionClosable, id, "true")
				time.Sleep(10 * time.Millisecond)
			}
		}
		if !closable && !*forceDeletion {
			continue
		}
		tabTitle := getTitleOfSession(sessionID)

		_, terminalRemovalErr := executeCmd(DbusPathSessions, DbusMethodTerminalRemoval, tID)
		if terminalRemovalErr != nil {
			color.Warn.Printf("Terminal with terminalId #%s can't be removed! %v\n", tID, terminalRemovalErr)
		} else {
			*didSomething = true
			color.Info.Printf("Closing terminal #%s with session #%s and title '%s'\n", tID, sessionID, *tabTitle)
		}
	}
}

// method to get terminal_ids of all open sessions
func getAllTerminalIDs() ([]string, error) {
	terminalIDOutput, err := executeCmd(DbusPathSessions, DbusMethodTerminalIDList)
	if err != nil {
		color.Error.Printf("Problem fetching terminalIDs of yakuake\n")
		return nil, err
	}
	var terminalIDs = strings.Split(terminalIDOutput, ",")
	return terminalIDs, nil
}

// get all session ids currently open
func getAllSessionIDs() ([]string, error) {
	sessionIDOutput, err := executeCmd(DbusPathSessions, DbusMethodSessionIDList)
	if err != nil {
		color.Error.Printf("Problem fetching sessionIDs of yakuake\n")
		return nil, err
	}
	var sessionIDs = strings.Split(sessionIDOutput, ",")
	return sessionIDs, nil
}

// get terminal ids of a single sessions id
func getTerminalIDsForSessionID(sessionID *string) []string {
	terminalIDOutput, _ := executeCmd(DbusPathSessions, DbusMethodTerminalIDListForSessionID, *sessionID)
	terminalIDs := strings.Split(terminalIDOutput, ",")
	return terminalIDs
}

// checks if yakuake window is shown
func isWindowShown() bool {
	isShownOutput, visibleErr := executeCmd(DbusPathMainwindow, DbusMethodQwidgetVisible)
	if visibleErr != nil {
		color.Error.Printf("Problem fetching open state of yakuake window. %v.\n", visibleErr)
		return true
	}
	parseBool, err := strconv.ParseBool(isShownOutput)
	if err != nil {
		color.Error.Printf("Problem parsing opening state of yakuake window. Value: '%s'. %v\n", isShownOutput, visibleErr)
		return false
	}
	return parseBool
}

// wrapper method to execute a command in a specific terminal
func executeCommandInTerminal(command string, terminalID string) {
	color.Info.Printf("Execute command '%s' in terminal #%s\n", command, terminalID)
	executeCmdVoid(DbusPathSessions, DbusMethodRunCommandInTerminal, terminalID, command)
}

// start new session (open a new tab) depending on split settings of this tab
func startSession(tab *TabDescription) *string {
	switch strings.ToLower(tab.SplitMode) {
	case "left-right", "horizontal", "lr":
		sessionID, _ := executeCmd(DbusPathSessions, DbusMethodAddSessionLr)
		return &sessionID
	case "top-bottom", "vertical", "tb":
		sessionID, _ := executeCmd(DbusPathSessions, DbusMethodAddSessionTb)
		return &sessionID
	case "quad", "qu":
		sessionID, _ := executeCmd(DbusPathSessions, DbusMethodAddSessionQu)
		return &sessionID
	default:
		// open single session
		sessionID, _ := executeCmd(DbusPathSessions, DbusMethodAddSession)
		return &sessionID
	}
}

// get tab title by session's id
func getTitleOfSession(sessionID string) *string {
	titleOutput, _ := executeCmd(DbusPathTabs, DbusMethodTabTitle, sessionID)
	return &titleOutput
}

// check if terminal and its sessions can be closed by this tool
func isTerminalClosable(terminalID string) (bool, *string) {
	sessionsIDOutput := getSessionIDForTerminalID(terminalID)
	sessionIDs := strings.Split(sessionsIDOutput, ",")
	for _, sessionID := range sessionIDs {
		title := getTitleOfSession(sessionID)
		out, err := executeCmd(DbusPathSessions, DbusMethodIsSessionClosable, sessionID)
		if err != nil {
			color.Error.Printf("Error fetching closable information of session '%s'\n", sessionID)
			return false, title
		}
		isClosable, _ := strconv.ParseBool(out)
		if !isClosable {
			return false, title
		}
	}
	return true, nil
}

// get session id by terminal id
func getSessionIDForTerminalID(terminalID string) string {
	output, err := executeCmd(DbusPathSessions, DbusMethodSessionIDForTerminalID, terminalID)
	if err != nil {
		color.Error.Printf("Error %v\n", err)
		return ""
	}
	return output
}

// method wrapper to execute all the dbus commands
func executeCmd(args ...string) (string, error) {
	arguments := append([]string{DbusService}, args...)
	cmd := exec.Command(DbusApp, arguments...)
	out, err := cmd.Output()
	if err != nil {
		color.Error.Println(err.Error())
		return "", err
	}
	return strings.Trim(string(out), "\n"), nil
}

// execute dbus cmd without using the output, errors are printed
func executeCmdVoid(args ...string) {
	_, err := executeCmd(args...)
	if err != nil {
		color.Warn.Println(err.Error())
		return
	}
	return
}
