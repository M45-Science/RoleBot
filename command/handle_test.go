package command

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name        string
		permissions int64
		requested   int64
		want        bool
	}{
		{"matching permission", discordgo.PermissionManageRoles, discordgo.PermissionManageRoles, true},
		{"administrator grants permission", discordgo.PermissionAdministrator, discordgo.PermissionManageRoles, true},
		{"unrelated permission is rejected", discordgo.PermissionManageMessages, discordgo.PermissionManageRoles, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{Member: &discordgo.Member{Permissions: tt.permissions}}}
			if got := hasPermission(i, tt.requested); got != tt.want {
				t.Fatalf("hasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}
