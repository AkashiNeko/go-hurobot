package cmds

import (
	"errors"
	"fmt"
	"strings"

	"go-hurobot/config"
	"go-hurobot/qbot"

	"github.com/gorcon/rcon"
	"gorm.io/gorm"
)

func cmd_mc(c *qbot.Client, raw *qbot.Message, args *ArgsList) {
	if args.Size < 2 {
		c.SendMsg(raw, "Usage: mc <command>")
		return
	}

	// Get RCON configuration for this group
	var rconConfig qbot.GroupRconConfigs
	result := qbot.PsqlDB.Where("group_id = ?", raw.GroupID).First(&rconConfig)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Silently return if RCON not configured for this group
			return
		}
		// Log other database errors if needed
		return
	}

	if !rconConfig.Enabled {
		c.SendMsg(raw, "RCON is disabled for this group")
		return
	}

	// Join all arguments after 'mc' as the command
	command := strings.Join(args.Contents[1:], " ")

	// Check permissions for non-master users
	if raw.UserID != config.MasterID && !isAllowedCommand(command) {
		c.SendMsg(raw, "Permission denied. You can only use query commands.")
		return
	}

	// Execute RCON command
	response, err := executeRconCommand(rconConfig.Address, rconConfig.Password, command)
	if err != nil {
		c.SendMsg(raw, fmt.Sprintf("RCON error: %s", err.Error()))
		return
	}

	// Send response back (limit to avoid spam)
	if len(response) > 1000 {
		response = response[:1000] + "... (truncated)"
	}

	if response == "" {
		response = "No output"
	}

	c.SendMsg(raw, qbot.CQReply(raw.UserID)+response)
}

func executeRconCommand(address, password, command string) (string, error) {
	// Connect to RCON server
	conn, err := rcon.Dial(address, password)
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Execute command
	response, err := conn.Execute(command)
	if err != nil {
		return "", fmt.Errorf("failed: %w", err)
	}

	return response, nil
}

// isAllowedCommand checks if a command is allowed for non-master users
func isAllowedCommand(command string) bool {
	// Remove leading slash if present
	command = strings.TrimPrefix(command, "/")

	// Split command into parts for analysis
	parts := strings.Fields(strings.ToLower(command))
	if len(parts) == 0 {
		return false
	}

	mainCmd := parts[0]

	// Allowed commands for non-master users (query/read-only commands)
	switch mainCmd {
	case "list":
		return true
	case "seed":
		return true
	case "version":
		return true
	case "data":
		// Only allow "data get" commands
		if len(parts) >= 2 && parts[1] == "get" {
			return true
		}
		return false
	case "team":
		// Only allow "team list"
		if len(parts) >= 2 && parts[1] == "list" {
			return true
		}
		return false
	case "whitelist":
		// Only allow "whitelist list"
		if len(parts) >= 2 && parts[1] == "list" {
			return true
		}
		return false
	case "banlist":
		return true
	case "locate":
		// Allow all locate subcommands (structure, biome, poi)
		return true
	case "worldborder":
		// Only allow "worldborder get"
		if len(parts) >= 2 && parts[1] == "get" {
			return true
		}
		return false
	case "datapack":
		// Only allow "datapack list"
		if len(parts) >= 2 && parts[1] == "list" {
			return true
		}
		return false
	case "function":
		// Allow function queries (without execution)
		// This is tricky - for safety, only allow when no arguments that suggest execution
		if len(parts) == 1 {
			return true // Just "function" command shows help
		}
		return false
	case "gamerule":
		// Allow gamerule queries (when no value is being set)
		if len(parts) <= 2 {
			return true // "gamerule" or "gamerule <rule>" (query)
		}
		return false // "gamerule <rule> <value>" (modification)
	case "difficulty":
		// Allow difficulty query (when no value is being set)
		if len(parts) == 1 {
			return true // Just "difficulty" (query)
		}
		return false // "difficulty <value>" (modification)
	case "defaultgamemode":
		// Allow defaultgamemode query (when no value is being set)
		if len(parts) == 1 {
			return true // Just "defaultgamemode" (query)
		}
		return false // "defaultgamemode <value>" (modification)
	default:
		return false
	}
}
