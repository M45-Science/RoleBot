package command

import (
	"RoleBot/cfg"
	"RoleBot/cwlog"
	"RoleBot/db"
	"RoleBot/disc"
	"RoleBot/glob"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Command func(s *discordgo.Session, i *discordgo.InteractionCreate, guild *db.GuildData)
	AppCmd  *discordgo.ApplicationCommand

	AdminOnly bool
	ModOnly   bool
}

var CL []Command

var adminPerms int64 = discordgo.PermissionAdministrator
var modPerms int64 = discordgo.PermissionManageRoles
var defaultPerms int64 = discordgo.PermissionUseSlashCommands

func hasPermission(i *discordgo.InteractionCreate, permission int64) bool {
	if i.Member == nil {
		return false
	}

	permissions := i.Member.Permissions
	return permissions&discordgo.PermissionAdministrator != 0 || permissions&permission != 0
}

func RegisterCommands(s *discordgo.Session) {
	CL = cmds

	for i, o := range CL {

		if o.AdminOnly {
			o.AppCmd.DefaultMemberPermissions = &adminPerms
		} else if o.ModOnly {
			o.AppCmd.DefaultMemberPermissions = &modPerms
		} else {
			o.AppCmd.DefaultMemberPermissions = &defaultPerms
		}

		cmd, err := s.ApplicationCommandCreate(cfg.Config.App, "", o.AppCmd)
		if err != nil {
			cwlog.DoLog("Failed to create command: " + CL[i].AppCmd.Name)
			continue
		} else {
			cwlog.DoLog("Registered command: " + CL[i].AppCmd.Name)
		}
		CL[i].AppCmd = cmd
	}
}

func ClearCommands() {
	if *glob.DoDeregisterCommands && disc.Session != nil {
		cmds, _ := disc.Session.ApplicationCommands(cfg.Config.App, "")
		for _, v := range cmds {
			cwlog.DoLog(fmt.Sprintf("Deregistered command: %s", v.Name))
			err := disc.Session.ApplicationCommandDelete(disc.Session.State.User.ID, "", v.ID)
			if err != nil {
				cwlog.DoLog(err.Error())
			}
		}
	}
}
