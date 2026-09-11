package command

import (
	"RoleBot/db"
	"RoleBot/disc"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var cmds = []Command{
	{
		AppCmd: &discordgo.ApplicationCommand{
			Name:        "get-roles",
			Description: "Sends drop-down menu: Assign or remove roles from yourself.",
		},
		Command: GetRoles,
	},
	{
		AppCmd: &discordgo.ApplicationCommand{
			Name:        "configure-bot",
			Description: "Sends drop-down menu: Add or remove roles to the list, users can self-assign/remove these.",
		},
		Command: ConfigureBot,
		ModOnly: true,
	},
}

func GetRoles(s *discordgo.Session, i *discordgo.InteractionCreate, guild *db.GuildData) {
	disc.EphemeralResponse(s, i, disc.DiscPurple, "Status:", "Finding eligible roles.", false)
	db.LookupRoleNames(s, guild)

	var availableRoles []discordgo.SelectMenuOption

	for _, role := range guild.Roles {

		found := false

		//Show roles already assigned to us
		userRoleNames := i.Member.Roles
		userRoles := []db.RoleData{}
		/* The list of roles from a specific user are just text, look up the actual role data. */
		/* TODO: Flip these and optimize */
		for _, guildRole := range guild.Roles {
			for _, userRole := range userRoleNames {
				if strings.EqualFold(userRole, db.IntToSnowflake(guildRole.ID)) {
					userRoles = append(userRoles, guildRole)
				}
			}
		}
		// Make component list.
		for _, existing := range userRoles {
			if existing.ID == role.ID {

				entry := discordgo.SelectMenuOption{
					Emoji: &discordgo.ComponentEmoji{
						Name: "✅",
					},
					Label: role.Name, Value: db.IntToSnowflake(role.ID)}
				availableRoles = append(availableRoles, entry)
				found = true
				break
			}
		}

		if !found {
			entry := discordgo.SelectMenuOption{Label: role.Name, Value: db.IntToSnowflake(role.ID)}
			availableRoles = append(availableRoles, entry)
		}

	}

	if len(availableRoles) <= 0 {
		embed := []*discordgo.MessageEmbed{{
			Title:       "Error:",
			Description: "Sorry, there are no roles available to you right now.",
		}}
		respose := &discordgo.WebhookEdit{
			Embeds: &embed,
		}

		s.InteractionResponseEdit(i.Interaction, respose)
		return
	}

	embed := []*discordgo.MessageEmbed{{
		Title:       "Info:",
		Description: "Select roles to add or remove them from yourself.\n\n✅ = Roles you already have.",
	}}
	respose := &discordgo.WebhookEdit{
		Components: &[]discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						// Select menu, as other components, must have a customID, so we set it to this value.
						CustomID:    "assign-roles",
						Placeholder: "Choose roles",
						Options:     availableRoles,
						MaxValues:   len(availableRoles),
					},
				},
			},
		},
		Embeds: &embed,
	}

	s.InteractionResponseEdit(i.Interaction, respose)

}

func ConfigureBot(s *discordgo.Session, i *discordgo.InteractionCreate, guild *db.GuildData) {

	/* Reply right away, so we don't time out looking up role names */
	disc.EphemeralResponse(s, i, disc.DiscPurple, "Status:", "Finding eligible roles.", false)

	var availableRoles []discordgo.SelectMenuOption
	roles := disc.GetGuildRoles(s, i.GuildID)

	/*
	 * Cycle through all guild roles, and make a list for the drop-down menu
	 * First add eligible roles from the guild that aren't in the database.
	 * Then add roles that are in the database, and mark them as selected with a green check
	 */
	for _, role := range roles {

		//Block specific names
		if strings.EqualFold(role.Name, "@everyone") {
			continue
		}

		found := false

		/* Exclude roles with moderator permissions */
		if role.Permissions&(discordgo.PermissionAdministrator|
			discordgo.PermissionBanMembers|
			discordgo.PermissionManageRoles|
			discordgo.PermissionModerateMembers|
			discordgo.PermissionManageWebhooks|
			discordgo.PermissionManageServer) != 0 {
			found = true
		}

		/* Skip roles already in database */
		for _, existing := range guild.Roles {
			if db.IntToSnowflake(existing.ID) == role.ID {
				found = true
				break
			}
		}

		/* Role is not already in the database */
		if !found {
			entry := discordgo.SelectMenuOption{Label: role.Name, Value: role.ID}
			availableRoles = append(availableRoles, entry)
		}

	}

	/*
	 * Show roles that are already selected, even if no longer existing
	 * This shouldnt happen, but would be helpful just in case it does
	 */
	for _, grole := range guild.Roles {
		entry := discordgo.SelectMenuOption{
			Emoji: &discordgo.ComponentEmoji{
				Name: "✅",
			},
			Label: grole.Name, Value: db.IntToSnowflake(grole.ID)}
		availableRoles = append(availableRoles, entry)
	}

	/* Let user know if there are no eligible roles. */
	if len(availableRoles) <= 0 {
		embed := []*discordgo.MessageEmbed{{
			Title:       "Error:",
			Description: "Sorry, there are no eligible roles!",
		}}
		respose := &discordgo.WebhookEdit{
			Embeds: &embed,
		}

		s.InteractionResponseEdit(i.Interaction, respose)
		return
	}

	embed := []*discordgo.MessageEmbed{{
		Title:       "Info:",
		Description: "Select roles to add them to the list.\nRoles in the list can be self-assigned by users.\n\n✅ = Already in list, selecting will remove.\nRoles with moderator permissions are excluded.",
	}}
	respose := &discordgo.WebhookEdit{
		Components: &[]discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						// Select menu, as other components, must have a customID, so we set it to this value.
						CustomID:    "config-roles",
						Placeholder: "Choose roles",
						Options:     availableRoles,
						MaxValues:   len(availableRoles),
					},
				},
			},
		},
		Embeds: &embed,
	}

	s.InteractionResponseEdit(i.Interaction, respose)
}
