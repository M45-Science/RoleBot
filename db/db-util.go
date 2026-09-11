package db

import (
	"RoleBot/cons"
	"RoleBot/cwlog"
	"RoleBot/disc"
	"bytes"
	"compress/zlib"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Current time in compact format
func NowToCompact() uint32 {
	tNow := time.Now().UTC().Unix()
	return uint32(tNow - cons.RoleBotEpoch)
}

// Compact format to unix time
func CompactToUnix(input uint32) uint64 {
	return uint64(input) + cons.RoleBotEpoch
}

func UnixToCompact(input uint64) uint32 {
	return uint32(input - cons.RoleBotEpoch)
}

func IntToSnowflake(id uint64) string {
	strId := fmt.Sprintf("%v", id)
	return strId
}

func SnowflakeToInt(i string) (uint64, error) {
	return strconv.ParseUint(i, 10, 64)
}

func GuildRoleUpdate(s *discordgo.Session, role *discordgo.GuildRoleUpdate) {

	guild := GuildLookupReadString(role.GuildID)
	if guild != nil {
		found := -1
		/* See if the role exists in the DB */
		for rpos, dbrole := range guild.Roles {
			if IntToSnowflake(dbrole.ID) == role.Role.ID {
				found = rpos
				break
			}
		}
		if found >= 0 {
			/* Update role name */
			cwlog.DoLog(fmt.Sprintf("Event: Updated role: %v for guild %v.", role.Role.ID, guild.Guild))
			guild.Roles[found].Name = role.Role.Name
		}
	}
}
func GuildRoleDelete(s *discordgo.Session, role *discordgo.GuildRoleDelete) {

	guild := GuildLookupReadString(role.GuildID)
	if guild != nil {
		found := -1
		/* See if the role exists in the DB */
		for rpos, dbrole := range guild.Roles {
			if IntToSnowflake(dbrole.ID) == role.RoleID {
				found = rpos
				break
			}
		}
		if found >= 0 {
			/* Delete role */
			cwlog.DoLog(fmt.Sprintf("Event: Removed role: %v for guild %v.", role.RoleID, guild.Guild))
			guild.Roles = append(guild.Roles[:found], guild.Roles[found+1:]...)
		}
	}
}

func compressZip(data []byte) []byte {
	var b bytes.Buffer
	w, err := zlib.NewWriterLevel(&b, zlib.BestSpeed)
	if err != nil {
		log.Println("ERROR: Gzip writer failure:", err)
	}
	w.Write(data)
	w.Close()
	return b.Bytes()
}

func LookupRoleNames(s *discordgo.Session, guildData *GuildData) {
	startTime := time.Now()
	count := 0

	//Process all guilds
	if guildData == nil {
		GuildLookupLock.RLock()
		for gpos, guild := range GuildLookup {
			time.Sleep(cons.LockRest)
			guild.Lock.Lock()

			for rpos, role := range guild.Roles {
				/* Only look up roles with no cache */
				if rpos < cons.LimitRoles && role.Name == "" {
					roleList := disc.GetGuildRoles(s, IntToSnowflake(guild.Guild))
					for _, discRole := range roleList {
						discRoleID, err := SnowflakeToInt(discRole.ID)
						if err == nil {
							if role.ID == discRoleID {
								if discRole.Name != "" {

									GuildLookup[gpos].Roles[rpos].Name = discRole.Name
									count++
								}
							}
						}
					}
				}
			}

			guild.Lock.Unlock()
		}
		GuildLookupLock.RUnlock()
		buf := fmt.Sprintf("Added %v role names in %v.", count, time.Since(startTime).String())
		cwlog.DoLog(buf)
	} else { //Process a specific guild

		guildData.Lock.Lock()
		for rpos, role := range guildData.Roles {
			discGuild, err := s.Guild(IntToSnowflake(guildData.Guild))
			if err == nil {
				for _, discRole := range discGuild.Roles {
					if IntToSnowflake(role.ID) == discRole.ID {
						guildData.Roles[rpos].Name = discRole.Name
					}
				}
			}
		}
		guildData.Lock.Unlock()
	}
}
