package cmds

import (
	"fmt"
	"strconv"
	"strings"

	"go-hurobot/config"
	"go-hurobot/qbot"
)

func cmd_rcon(c *qbot.Client, raw *qbot.Message, args *ArgsList) {
	if raw.UserID != config.MasterID {
		return
	}

	const help = `Usage: rcon [status | set <address> <password> | enable | disable]

Examples:
  rcon status
  rcon set '127.0.0.1:25575' 'password'
  rcon enable
  rcon disable`

	if args.Size == 1 {
		c.SendMsg(raw, help)
		return
	}

	switch args.Contents[1] {
	case "status":
		showRconStatus(c, raw)
	case "set":
		if args.Size != 4 {
			c.SendMsg(raw, "Usage: rcon set <address> <password>")
			return
		}
		setRconConfig(c, raw, args.Contents[2], args.Contents[3])
	case "enable":
		toggleRcon(c, raw, true)
	case "disable":
		toggleRcon(c, raw, false)
	default:
		c.SendMsg(raw, help)
	}
}

func showRconStatus(c *qbot.Client, msg *qbot.Message) {
	var config qbot.GroupRconConfigs
	result := qbot.PsqlDB.Where("group_id = ?", msg.GroupID).First(&config)

	if result.Error != nil {
		c.SendMsg(msg, "RCON not configured for this group")
		return
	}

	status := "disabled"
	if config.Enabled {
		status = "enabled"
	}

	// Hide password for security
	maskedPassword := strings.Repeat("*", len(config.Password))
	response := fmt.Sprintf("RCON Status: %s\nAddress: %s\nPassword: %s",
		status, config.Address, maskedPassword)

	c.SendMsg(msg, response)
}

func setRconConfig(c *qbot.Client, msg *qbot.Message, address, password string) {
	// Validate address format (should contain port)
	if !strings.Contains(address, ":") {
		c.SendMsg(msg, "Invalid address format. Use host:port (e.g., 127.0.0.1:25575)")
		return
	}

	// Validate port
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		c.SendMsg(msg, "Invalid address format. Use host:port")
		return
	}

	if port, err := strconv.Atoi(parts[1]); err != nil || port < 1 || port > 65535 {
		c.SendMsg(msg, "Invalid port number")
		return
	}

	config := qbot.GroupRconConfigs{
		GroupID:  msg.GroupID,
		Address:  address,
		Password: password,
		Enabled:  false, // Default to disabled for security
	}

	// Use Upsert to create or update
	result := qbot.PsqlDB.Where("group_id = ?", msg.GroupID).Assign(
		qbot.GroupRconConfigs{
			Address:  address,
			Password: password,
		},
	).FirstOrCreate(&config)

	if result.Error != nil {
		c.SendMsg(msg, "Database error: "+result.Error.Error())
		return
	}

	c.SendMsg(msg, fmt.Sprintf("RCON configuration updated:\nAddress: %s\nStatus: disabled (use 'rcon enable' to enable)", address))
}

func toggleRcon(c *qbot.Client, msg *qbot.Message, enabled bool) {
	// Check if configuration exists
	var config qbot.GroupRconConfigs
	result := qbot.PsqlDB.Where("group_id = ?", msg.GroupID).First(&config)

	if result.Error != nil {
		c.SendMsg(msg, "RCON not configured for this group. Use 'rcon set' first.")
		return
	}

	// Update enabled status
	result = qbot.PsqlDB.Model(&config).Where("group_id = ?", msg.GroupID).Update("enabled", enabled)

	if result.Error != nil {
		c.SendMsg(msg, "Database error: "+result.Error.Error())
		return
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}
	c.SendMsg(msg, fmt.Sprintf("RCON %s for this group", status))
}
