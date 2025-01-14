package main

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/akkuman/webeye"
	"github.com/akkuman/webeye/finger"
	"github.com/akkuman/webeye/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/remeh/sizedwaitgroup"
	"github.com/urfave/cli/v3"
)

func LoadFinger(templateFileURL string) (*finger.WebFingerSystem, error) {
	content, err := utils.GetLocalFileOrWeb(templateFileURL)
	if err != nil {
		return nil, err
	}
	return finger.ParseWebFinger(string(content))
}

func LoadTarget(targetListFileURL string) (targets []string, err error) {
	content, err := utils.GetLocalFileOrWeb(targetListFileURL)
	if err != nil {
		return nil, err
	}
	for _, t := range strings.Split(string(content), "\n") {
		targets = append(targets, strings.TrimSpace(t))
	}
	return targets, nil
}

func main() {
    cmd := &cli.Command{
		Usage: "get web app's fingerprint",
		Commands: []*cli.Command{
			{
				Name:  "findout",
				Usage: "find out fingerprint of website",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "template",
						Value: "https://raw.githubusercontent.com/0x727/FingerprintHub/refs/heads/main/web_fingerprint_v3.json",
						Usage: "template for fingerprint (format: FingerprintHub V3, from: filepath or url(with http(s)://))",
					},
					&cli.StringFlag{
						Name: "target-list",
						Usage: "target list file (from: filepath or url(with http(s)://)))",
						Required: true,
					},
					&cli.IntFlag{
						Name: "threads",
						Value: 0,
						Usage: "how many targets can be scanned simultaneously, default unlimited",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					wfs, err := LoadFinger(cmd.String("template"))
					if err != nil {
						return err
					}
					targets, err := LoadTarget(cmd.String("target-list"))
					if err != nil {
						return err
					}
					table := tablewriter.NewWriter(os.Stdout)
					table.SetHeader([]string{"target", "finger", "error"})
					rowCh := make(chan []string, 10)
					swg := sizedwaitgroup.New(int(cmd.Int("threads")))
					for _, target := range targets {
						swg.Add()
						go func()  {
							defer swg.Done()
							res, err := webeye.GetWebFinger(context.Background(), target, *wfs)
							var targetFingers []string
							for _, r := range res {
								targetFingers = append(targetFingers, r.Name)
							}
							var errText string
							if err != nil {
								errText = err.Error()
							}
							rowCh <- []string{target, strings.Join(targetFingers, ","), errText}
						}()
					}
					var finishWG sync.WaitGroup
					finishWG.Add(1)
					go func()  {
						defer finishWG.Done()
						for row := range rowCh {
							table.Append(row)
							table.Render()
						}
					}()
					swg.Wait()
					close(rowCh)
					finishWG.Wait()
					return nil
				},
			},
		},
    }

    if err := cmd.Run(context.Background(), os.Args); err != nil {
        log.Fatal(err)
    }
}