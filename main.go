package main

import (
	"RoleBot/cfg"
	"RoleBot/command"
	"RoleBot/cons"
	"RoleBot/cwlog"
	"RoleBot/db"
	"RoleBot/disc"
	"RoleBot/glob"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

func main() {
	/* Make data directory */
	errr := os.MkdirAll("data", os.ModePerm)
	if errr != nil {
		fmt.Print(errr.Error())
		return
	}
	/* Make log directory */
	errr = os.MkdirAll("data/db", os.ModePerm)
	if errr != nil {
		fmt.Print(errr.Error())
		return
	}

	glob.ServerRunning = true
	glob.DoRegisterCommands = flag.Bool("regCommands", false, "Register discord commands")
	glob.DoDeregisterCommands = flag.Bool("deregCommands", false, "Deregister discord commands")
	flag.Parse()

	glob.Uptime = time.Now().UTC().Round(time.Second)
	cwlog.StartLog()

	if !cfg.ReadCfg() {
		cwlog.DoLog("Configuration is invalid; refusing to overwrite it.")
		return
	}
	if !cfg.WriteCfg() {
		cwlog.DoLog("Unable to write configuration.")
		return
	}

	db.GuildLookup = make(map[uint64]*db.GuildData, cons.TSize)

	if !db.ReadAllClusters() {
		cwlog.DoLog("Database initialization failed.")
		return
	}
	db.UpdateGuildLookup()

	go startbot()

	/* Wait here for process signals */
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	tNow := time.Now().Round(time.Second)
	cwlog.DoLog("uptime: " + tNow.Sub(glob.Uptime).String())
	command.ClearCommands()
}

var DiscordConnectAttempts int

func startbot() {
	if cfg.Config.Token == "" {
		cwlog.DoLog("No discord token.")
		return
	}

	cwlog.DoLog(cons.BotName + " " + cons.Version + " starting.")
	cwlog.DoLog("Max Guilds: " + strconv.FormatInt(cons.MaxGuilds, 10))

	bot, err := discordgo.New("Bot " + cfg.Config.Token)

	if err != nil {
		cwlog.DoLog(fmt.Sprintf("An error occurred when attempting to create the Discord session. Details: %v", err))
		time.Sleep(time.Minute * (5 * cons.MaxDiscordAttempts))
		DiscordConnectAttempts++

		if DiscordConnectAttempts < cons.MaxDiscordAttempts {
			startbot()
		}
		return
	}

	bot.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)

	bot.AddHandler(botReady)
	//bot.AddHandler(db.GuildRoleCreate)
	bot.AddHandler(db.GuildRoleUpdate)
	bot.AddHandler(db.GuildRoleDelete)

	errb := bot.Open()

	if errb != nil {
		cwlog.DoLog(fmt.Sprintf("An error occurred when attempting to create the Discord session. Details: %v", errb))
		time.Sleep(time.Minute * (5 * cons.MaxDiscordAttempts))
		DiscordConnectAttempts++

		if DiscordConnectAttempts < cons.MaxDiscordAttempts {
			startbot()
		}
		return
	}

	bot.LogLevel = discordgo.LogWarning
}

func botReady(s *discordgo.Session, r *discordgo.Ready) {

	botstatus := cfg.Config.Domain
	err := s.UpdateGameStatus(0, botstatus)
	if err != nil {
		cwlog.DoLog(err.Error())
	}

	s.AddHandler(command.SlashCommand)

	disc.Session = s
	disc.Ready = r
	cwlog.DoLog("Discord bot ready")

	go db.LookupRoleNames(s, nil)

	if *glob.DoRegisterCommands {
		command.RegisterCommands(s)
	}
	if *glob.DoDeregisterCommands {
		command.ClearCommands()
	}

	go MainLoop()

}
