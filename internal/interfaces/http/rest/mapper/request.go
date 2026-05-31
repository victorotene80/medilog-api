package mapper

import (
	"strings"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/request"
)

func LoginRequestToCommand(req request.LoginRequest) command.LoginCommand {
	cmd := command.LoginCommand{
		Password: req.Password,
	}

	if email := strings.TrimSpace(req.Email); email != "" {
		cmd.Email = &email
	}

	if phone := strings.TrimSpace(req.Phone); phone != "" {
		cmd.Phone = &phone
	}

	return cmd
}
