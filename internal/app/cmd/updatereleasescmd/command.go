package updatereleasescmd

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bitmagnet-io/bitmagnet/internal/boilerplate/lazy"
	"github.com/bitmagnet-io/bitmagnet/internal/database/dao"
	"github.com/bitmagnet-io/bitmagnet/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

type Params struct {
	fx.In
	Dao    lazy.Lazy[*dao.Query]
	Logger *zap.SugaredLogger
}

type Result struct {
	fx.Out
	Command *cli.Command `group:"commands"`
}

type GithubFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	Url         string `json:"url"`
	DownloadUrl string `json:"download_url"`
}

func New(p Params) (Result, error) {
	return Result{Command: &cli.Command{
		Name:  "update_releases",
		Usage: "Fetch new releases",
		Flags: []cli.Flag{},
		Action: func(ctx *cli.Context) error {

			d, err := p.Dao.Get()
			if err != nil {
				return err
			}

			// TODO: Fetch most recent releases from https://api.predb.net/

			var lastImportTime time.Time
			result := d.Release.WithContext(ctx.Context).UnderlyingDB().Table("releases").Select("max(nzedbpre_dump)")
			err = result.Row().Scan(&lastImportTime)
			if err != nil {
				lastImportTime = time.Unix(0, 0)
			} else {
				err = result.Row().Scan(&lastImportTime)

				if err != nil {
					return err
				}
			}
			fmt.Printf("Fetching list of predb dumps since %s... \n", lastImportTime.String())

			var allDumps map[int64]GithubFile = make(map[int64]GithubFile)
			var lastImport int64 = lastImportTime.Unix()
			var nextDumps []GithubFile
			var folderTimestamps [2]int64
			var folders [2]GithubFile
			folderTimestamps[0] = math.MinInt
			folderTimestamps[1] = math.MaxInt

			// Create a Resty Client
			client := resty.New()
			url := "https://api.github.com/repos/nZEDb/nZEDbPre_Dumps/contents/dumps/"
			resp, err := client.R().
				Get(url)
			if err != nil {
				fmt.Println("Can't download folder listing:", err)
				os.Exit(1)
			} else {
				var files []GithubFile
				err := json.Unmarshal(resp.Body(), &files)
				// figure out which folders to open
				if err != nil {
					fmt.Println("Cant unmarshal folder listing: ", err)
					os.Exit(1)
				} else {
					for _, folder := range files {
						if folder.Type == "dir" && strings.HasPrefix(folder.Path, "dumps/") {
							folderDate, err := strconv.ParseInt(folder.Name, 10, 64)
							if err != nil {
								fmt.Println("Error: Couldn't convert folder name to timestamp", err)
								os.Exit(1)
							} else {
								if folderDate < lastImport && folderDate > folderTimestamps[0] { // lower folder
									folderTimestamps[0] = folderDate
									folders[0] = folder
								}
								if folderDate > lastImport && folderDate < folderTimestamps[1] { // upper folder
									folderTimestamps[1] = folderDate
									folders[1] = folder
								}
							}
						}
					}
					// open folders with dumps
					for _, folder := range folders {
						if folder.Url == "" {
							continue
						}
						resp, err := client.R().
							Get(folder.Url)
						if err != nil {
							fmt.Println("Error: Can't download dump file listing ", err)
							os.Exit(1)
						} else {
							var dumps []GithubFile
							err := json.Unmarshal(resp.Body(), &dumps)
							if err != nil {
								fmt.Println("Error: Can't Unmarshal dup file listing - ", err)
								os.Exit(1)
							} else {
								for _, dump := range dumps {
									dumpDate, err := strconv.ParseInt(dump.Name[:strings.IndexByte(dump.Name, '_')], 10, 64)
									if err != nil {
										fmt.Println("Error: dump filename issue", err)
										os.Exit(1)
									} else {
										allDumps[dumpDate] = dump
									}
								}
							}

						}
					}
				}
			}
			// sort the dumps and find the next one to import
			keys := make([]int64, 0, len(allDumps))
			for k := range allDumps {
				keys = append(keys, k)
			}

			// add up to next 10 dumps to list of imports
			// Note: Github API is rate limited to 60 req/h, but downloads aren't so this batch size
			// could be set higher or even unlimited to download full folders at once
			batchSize := 50
			sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
			for idx, k := range keys {
				if k > lastImport && len(nextDumps) == 0 {
					for i := 0; i < batchSize; i++ {
						if idx+i < len(keys) {
							nextDumps = append(nextDumps, allDumps[keys[idx+i]])
						}
					}
				}
			}

			// import dump if there's a newer one

			loc, _ := time.LoadLocation("Europe/Berlin") // this isn't explicit but it seems that the dumps are created with CET/CEST timestamps
			for _, nextDump := range nextDumps {
				resp, err := client.R().Get(nextDump.DownloadUrl)
				if err != nil {
					fmt.Println("Error downloading dump:", err)
					os.Exit(1)
				} else {

					releases := make([]model.Release, 0)
					dumpDate, _ := strconv.ParseInt(nextDump.Name[:strings.IndexByte(nextDump.Name, '_')], 10, 64) // no need to check error, already done above
					fmt.Printf("Importing %s from %s, size: %d \n", nextDump.Path, time.Unix(dumpDate, 0).String(), len(resp.Body()))
					gzreader, err := gzip.NewReader(bytes.NewReader([]byte(resp.Body())))
					if err != nil {
						fmt.Println("can't unpack dump", err)
						os.Exit(1)
					}

					buf := new(bytes.Buffer)
					buf.ReadFrom(gzreader)
					dumpContents := buf.String()

					lines := strings.Split(dumpContents, "\r\n")
					for _, line := range lines {
						if len(line) == 0 {
							continue // skip empty lines
						}
						fields := strings.Split(line, "\t\t")
						for idx, field := range fields {
							if field == "\\N" {
								// replace \N
								fields[idx] = ""
							} else {
								// strip quotes
								fields[idx] = field[1 : len(field)-1]
							}
							//fmt.Print(field, ",")
						}
						//fmt.Println("----") // Println will add back the final '\n' */
						createdTime, err := time.ParseInLocation("2006-01-02 15:04:05", fields[8], loc)
						if err != nil {
							println("cant parse date")
							panic(err)
						}
						nuked, err := strconv.ParseInt(fields[5], 10, 32)
						if err != nil {
							println("cant parse int")
							panic(err)
						}
						release := model.Release{
							Title:        fields[0],
							Nfo:          model.NewNullString(fields[1]),
							Size:         model.NewNullString(fields[2]),
							Files:        model.NewNullString(fields[3]),
							Filename:     model.NewNullString(fields[4]),
							Nuked:        int32(nuked),
							Nukereason:   model.NewNullString(fields[6]),
							Category:     model.NewNullString(fields[7]),
							Created:      createdTime,
							Source:       model.NewNullString(fields[9]),
							Requestid:    model.NewNullString(fields[10]),
							Groupname:    model.NewNullString(fields[11]),
							NzedbpreDump: time.Unix(dumpDate, 0),
						}
						releases = append(releases, release)
					}
					t := d.Release
					result := t.UnderlyingDB().Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&releases, 1000) // Create in bulk to make sure we get the whole dump in one transaction, ignore lines that were imported fron other dumps
					if result.Error != nil {
						panic(result.Error)
					}
				}
			}
			return nil
		},
	}}, nil
}
