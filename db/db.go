package db

import (
	"RoleBot/cons"
	"RoleBot/cwlog"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"
)

func WriteLIDTop() {
	buf := fmt.Sprintf("LIDTop: %v", LID_TOP)
	err := os.WriteFile("data/db/"+cons.LIDTopFile+".tmp", []byte(buf), 0644)

	if err != nil && err != fs.ErrNotExist {
		cwlog.DoLog(err.Error())
		return
	}
	err = os.Rename("data/db/"+cons.LIDTopFile+".tmp", "data/db/"+cons.LIDTopFile)

	if err != nil {
		cwlog.DoLog("WriteLIDTop: Couldn't rename file: " + cons.LIDTopFile)
		return
	}
}

func ReadLIDTop() int64 {
	data, err := os.ReadFile("data/db/" + cons.LIDTopFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		cwlog.DoLog("Unable to read LIDTop file: " + err.Error())
		return -1
	}
	splitData := strings.Split(string(data), ":")

	if len(splitData) > 1 {
		valString := splitData[1]
		cleanVal := strings.TrimSuffix(valString, "\n")
		cleanVal = strings.TrimPrefix(cleanVal, " ")
		intVal, err := strconv.ParseUint(cleanVal, 10, 32)

		if err != nil {
			cwlog.DoLog("Unable to parse LIDTop value: " + err.Error())
			return -1
		}

		return int64(intVal)
	}
	return -1
}

func WriteAllCluster() {

	startTime := time.Now()

	for i := 0; i < cons.NumClusters; i++ {
		WriteCluster(i)
	}
	endTime := time.Now()
	cwlog.DoLog("DB Write Complete, took: " + endTime.Sub(startTime).String())

}

const VERSION_size = 2

const LID_size = 4
const SNOWFLAKE_size = 8
const ADDED_size = 4
const DONOR_size = 1
const NUMROLE_size = 2

const MAXROLES = 0xFFFF
const MAXROLE_size = SNOWFLAKE_size * MAXROLES

const staticSize = LID_size +
	SNOWFLAKE_size +
	ADDED_size +
	DONOR_size +
	NUMROLE_size

func WriteCluster(i int) {
	maxGuildsPerCluster := cons.MaxGuilds / cons.NumClusters
	maxPersistedRoles := min(cons.LimitRoles, MAXROLES)
	buf := make([]byte, 0, VERSION_size+(staticSize+SNOWFLAKE_size*maxPersistedRoles)*maxGuildsPerCluster)
	buf = binary.LittleEndian.AppendUint16(buf, 1) // version number

	start := i*maxGuildsPerCluster + 1
	end := start + maxGuildsPerCluster
	for x := start; x < end; x++ {

		g := Database[x]
		if g == nil {
			break
		}

		Database[x].Lock.RLock()
		buf = binary.LittleEndian.AppendUint32(buf, g.LID)
		buf = binary.LittleEndian.AppendUint64(buf, g.Guild)
		buf = binary.LittleEndian.AppendUint32(buf, g.Added)
		buf = append(buf, g.Donator)

		numRoles := min(len(g.Roles), MAXROLES)
		buf = binary.LittleEndian.AppendUint16(buf, uint16(numRoles))

		for _, role := range g.Roles[:numRoles] {
			buf = binary.LittleEndian.AppendUint64(buf, role.ID)
		}

		Database[x].Lock.RUnlock()
	}
	/* Don't write empty files */
	if len(buf) > VERSION_size {
		name := fmt.Sprintf("data/db/cluster-%v.dat", i+1)
		err := os.WriteFile(name+".tmp", buf, 0644)
		if err != nil && err != fs.ErrNotExist {
			cwlog.DoLog(err.Error())
			return
		}
		err = os.Rename(name+".tmp", name)

		if err != nil {
			cwlog.DoLog("WriteCluster: Couldn't rename file: " + name)
			return
		}
	}
}

func ReadAllClusters() bool {

	startTime := time.Now()
	for x := 0; x < cons.NumClusters; x++ {
		ReadCluster(int64(x))
	}
	endTime := time.Now()

	lid := ReadLIDTop()
	if lid < 0 || lid != int64(LID_TOP) {
		cwlog.DoLog(fmt.Sprintf("LIDTop mismatch: file=%v, database=%v", lid, LID_TOP))
		return false
	}
	cwlog.DoLog("Read all clusters, took: " + endTime.Sub(startTime).String())
	return true
}

func ReadCluster(i int64) {
	startTime := time.Now()

	name := fmt.Sprintf("data/db/cluster-%v.dat", i+1)
	data, err := os.ReadFile(name)
	if err != nil {
		return
	}

	dataLen := int64(len(data))
	var b int64

	if (dataLen) < VERSION_size+staticSize {
		fmt.Println("cluster size too small:", dataLen, " bytes of ", VERSION_size+staticSize)
		return
	}

	version := binary.LittleEndian.Uint16(data[b:])
	b += 2
	if version == 1 {
		for b < dataLen {
			if dataLen-b < staticSize {
				cwlog.DoLog("Cluster record is truncated: " + name)
				return
			}

			g := new(GuildData)

			g.LID = binary.LittleEndian.Uint32(data[b:])
			b += 4
			g.Guild = binary.LittleEndian.Uint64(data[b:])
			b += 8
			g.Added = binary.LittleEndian.Uint32(data[b:])
			b += 4
			g.Donator = data[b]
			b += 1

			numRoles := binary.LittleEndian.Uint16(data[b:])
			b += 2
			rolesSize := int64(numRoles) * SNOWFLAKE_size
			if dataLen-b < rolesSize {
				cwlog.DoLog("Cluster roles are truncated: " + name)
				return
			}

			if numRoles == 0 {
				if g.LID >= cons.MaxGuilds {
					cwlog.DoLog("LID larger than maxguild.")
					break
				}
				LID_TOP++
				Database[LID_TOP] = g
				/* Found a role instead of record end */
			} else {
				roleData := []RoleData{}
				for i := uint16(1); i <= numRoles; i++ {
					roleID := binary.LittleEndian.Uint64(data[b:])
					b += 8

					roleData = append(roleData, RoleData{ID: roleID})
				}
				g.Roles = roleData
				LID_TOP++
				Database[LID_TOP] = g
			}
		}
	} else {
		cwlog.DoLog("Invalid cluster version.")
		return
	}

	endTime := time.Now()
	if b > 2 {
		cwlog.DoLog("Cluster-" + strconv.FormatInt(int64(i+1), 10) + " read, took: " + endTime.Sub(startTime).String() + ", Read: " + strconv.FormatInt(b, 10) + "b")
	}
}

func AppendCluster(guild *GuildData, cid uint32, gid uint32) {

}

func UpdateGuildLookup() {
	startTime := time.Now()

	GuildLookupLock.Lock()
	var x uint32
	for x = 1; x <= LID_TOP; x++ {
		if Database[x] != nil {
			GuildLookup[Database[x].Guild] = Database[x]
		}
	}
	GuildLookupLock.Unlock()

	endTime := time.Now()
	cwlog.DoLog("Guild lookup map update, took: " + endTime.Sub(startTime).String())

	WriteAllCluster()
	DumpGuilds()
}

func GuildLookupRead(i uint64) *GuildData {
	GuildLookupLock.RLock()
	g := GuildLookup[i]
	GuildLookupLock.RUnlock()
	return g
}

func GuildLookupReadString(i string) *GuildData {
	GuildLookupLock.RLock()
	val, err := SnowflakeToInt(i)
	if err == nil {
		g := GuildLookup[val]
		GuildLookupLock.RUnlock()
		return g
	}
	GuildLookupLock.RUnlock()
	return nil
}

// Does not need a lock, this is pre-allocated.
func AddGuild(guildid uint64) {
	cwlog.DoLog(fmt.Sprintf("AddGuild: %v", guildid))

	LID_TOP++
	tNow := NowToCompact()
	newGuild := GuildData{LID: LID_TOP, Guild: guildid, Added: uint32(tNow), Modified: uint32(tNow), Donator: 0}
	Database[LID_TOP] = &newGuild

	WriteLIDTop()
	UpdateGuildLookup()
}

func DumpGuilds() {
	GuildLookupLock.RLock()

	cwlog.DoLog("DumpGuilds: Writing guilds...")

	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")
	if err := enc.Encode(GuildLookup); err != nil {
		cwlog.DoLog("DumpGuilds: enc.Encode failure")
		GuildLookupLock.RUnlock()
		return
	}
	GuildLookupLock.RUnlock()

	nfilename := "data/" + cons.DumpName + ".tmp"
	//compBuf := compressZip(outbuf.Bytes())
	err := os.WriteFile(nfilename, outbuf.Bytes(), 0644)

	if err != nil {
		cwlog.DoLog("DumpGuilds: Couldn't write db temp file.")
		return
	}

	oldName := nfilename
	newName := "data/" + cons.DumpName
	err = os.Rename(oldName, newName)

	if err != nil {
		cwlog.DoLog("DumpGuilds: Couldn't rename db temp file.")
		return
	}

	cwlog.DoLog("DumpGuilds: Complete!")
}
