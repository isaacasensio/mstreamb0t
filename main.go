package main

import (
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/genuinetools/pkg/cli"
	"github.com/isaacasensio/mstreamb0t/mstream"
	"github.com/isaacasensio/mstreamb0t/pbnotifier"
	"github.com/sirupsen/logrus"
	pushbullet "github.com/xconstruct/go-pushbullet"
)

var (
	mangaNames string
	token      string

	interval time.Duration
	once     bool
)

// Config contains cli config
type Config struct {
}

// GetConfigDirPath returns path to config folder
func (c Config) GetConfigDirPath() (string, error) {
	// Get home directory.
	home := os.Getenv("HOME")
	if home != "" {
		return filepath.Join(home, ".mstreamb0t"), nil
	}
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(u.HomeDir, ".mstreamb0t"), nil
}

func main() {

	var config Config

	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "mstreamb0t"
	p.Description = "Bot that notifies you when a specified manga is released on MangaStream"

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)

	p.FlagSet.StringVar(&mangaNames, "manga-names", os.Getenv("MANGA_NAMES"), "Manga names")
	p.FlagSet.DurationVar(&interval, "interval", time.Minute, "update interval (ex. 5ms, 10s, 1m, 3h)")
	p.FlagSet.BoolVar(&once, "once", false, "run once and exit, do not run as a daemon")

	// Set the before function.
	p.Before = func(ctx context.Context) error {

		if len(mangaNames) < 1 {
			return errors.New("manga name cannot be empty")
		}

		configDir, err := config.GetConfigDirPath()
		if err != nil {
			return err
		}

		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			err = os.Mkdir(configDir, os.ModePerm)
			if err != nil {
				return err
			}
		}

		token := os.Getenv("PUSHBULLET_TOKEN")
		if token == "" {
			return errors.New("PUSHBULLET_TOKEN not found")
		}

		return nil
	}

	// Set the main program action.
	p.Action = func(ctx context.Context, args []string) error {
		ticker := time.NewTicker(interval)

		// On ^C, or SIGTERM handle exit.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		var cancel context.CancelFunc
		_, cancel = context.WithCancel(ctx)
		go func() {
			for sig := range c {
				cancel()
				ticker.Stop()
				logrus.Infof("Received %s, exiting.", sig.String())
				os.Exit(0)
			}
		}()

		configDir, _ := config.GetConfigDirPath()
		lastUpdatePath := path.Join(configDir, ".lastUpdate")
		mangas := strings.Split(mangaNames, ",")

		// If the user passed the once flag, just do the run once and exit.
		if once {
			run(mangas, getLastRun(lastUpdatePath))
			logrus.Info("Finished checking mangastream updates")
			updateLastRun(lastUpdatePath)
			os.Exit(0)
		}

		logrus.Infof("Starting bot to notify on mangastream updates every %s", interval)
		for range ticker.C {
			run(mangas, getLastRun(lastUpdatePath))
			updateLastRun(lastUpdatePath)
		}

		return nil
	}

	// Run our program.
	p.Run()
}

func run(mangas []string, lastUpdate time.Time) {

	mstream := mstream.NewClient("https://readms.net/rss")

	updates, err := mstream.FindNewReleasesSince(mangas, lastUpdate)
	if err != nil {
		panic(err)
	}

	if len(updates) == 0 {
		return
	}

	for _, item := range updates {
		logrus.Infof("New release: %s ", item)
	}

	token := os.Getenv("PUSHBULLET_TOKEN")
	pb := pbnotifier.NewClient(token, pushbullet.EndpointURL)
	err = pb.Notify("New manga(s) released!", strings.Join(updates, "\n"))
	if err != nil {
		panic(err)
	}

}

func getLastRun(lastUpdatePath string) time.Time {
	if _, err := os.Stat(lastUpdatePath); os.IsNotExist(err) {
		logrus.Info("First time running the script!")
		return time.Now().AddDate(-1, 0, 0)
	}
	b, err := ioutil.ReadFile(lastUpdatePath)
	if err != nil {
		panic(err)
	}
	time, err := time.Parse(time.RFC3339, string(b))
	if err != nil {
		panic(err)
	}
	logrus.Infof("Last time script was ran: %s", string(b))
	return time
}

func updateLastRun(lastUpdatePath string) {
	timeLastRun := time.Now().Format(time.RFC3339)
	err := ioutil.WriteFile(lastUpdatePath, []byte(timeLastRun), 0644)
	if err != nil {
		panic(err)
	}
}
