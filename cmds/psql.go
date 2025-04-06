package cmds

import (
	"database/sql"
	"fmt"
	"go-hurobot/config"
	"go-hurobot/qbot"
	"strings"
)

func cmd_psql(c *qbot.Client, raw *qbot.Message, args *ArgsList) {
	if raw.UserID != config.MasterID {
		c.SendMsg(raw, fmt.Sprintf("%s: Permission denied", args.Contents[0]))
		return
	}

	rows, err := qbot.PsqlDB.Raw(decodeSpecialChars(raw.Raw[5:])).Rows()
	if err != nil {
		c.SendMsg(raw, err.Error())
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		c.SendMsg(raw, err.Error())
		return
	}

	result := ""
	count := 1
	for rows.Next() {
		if count == 10 {
			result += "\n\n** more... **"
			break
		}

		values := make([]any, len(columns))
		for i := range values {
			values[i] = new(sql.RawBytes)
		}

		if err := rows.Scan(values...); err != nil {
			c.SendMsg(raw, err.Error())
			return
		}

		var rowStrings []string
		for i, col := range values {
			rowStrings = append(rowStrings, fmt.Sprintf("%s: %s", columns[i], string(*col.(*sql.RawBytes))))
		}

		if result != "" {
			result += "\n\n"
		}
		result += fmt.Sprintf("** %d **\n", count)
		result += strings.Join(rowStrings, "\n")
		count++
	}
	if err = rows.Err(); err != nil {
		c.SendMsg(raw, err.Error())
	} else if result == "" {
		c.SendMsg(raw, "[]")
	} else {
		c.SendMsg(raw, result)
	}
}
