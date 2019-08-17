package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Phantas0s/devdash/internal"
	"github.com/Phantas0s/devdash/internal/plateform"
	"golang.org/x/crypto/ssh/terminal"
)

// debug mode
var debug *bool

func main() {
	file := flag.String("config", ".devdash.yml", "The config file")
	debug = flag.Bool("debug", false, "Debug mode")
	term := flag.Bool("term", false, "Display terminal dimensions")
	flag.Parse()

	if *term {
		width, height, _ := terminal.GetSize(0)
		fmt.Printf("Width: %d, Height: %d", width, height)
		return
	}

	cfg, tui, err := loadFile(*file)
	if err != nil {
		internal.DisplayError(tui, err)
	}

	defer tui.Close()

	run(cfg.Projects, tui)

	if _, err := os.Stat(*file); os.IsNotExist(err) {
		internal.DisplayNoFile(tui)
		err := tui.AddCol("5")
		if err != nil {
			fmt.Println(err)
		}

		tui.AddRow()
		tui.Render()
	} else {
		ticker := time.NewTicker(time.Duration(cfg.RefreshTime()) * time.Second)
		go func() {
			for range ticker.C {
				tui.Clean()
				run(cfg.Projects, tui)
			}
		}()
	}

	tui.Loop()
}

func loadFile(file string) (config, *internal.Tui, error) {
	termui, err := plateform.NewTermUI(*debug)
	if err != nil {
		return config{}, nil, err
	}

	tui := internal.NewTUI(termui)
	data, _ := ioutil.ReadFile(file)
	cfg := mapConfig(data)
	tui.AddKQuit(cfg.KQuit())

	return cfg, tui, nil
}

func run(projects []Project, tui *internal.Tui) {
	for _, p := range projects {
		rows, sizes := p.OrderWidgets()
		project := internal.NewProject(p.Name, p.TitleOptions, rows, sizes, p.Themes)

		gaService := p.Services.GoogleAnalytics
		if !gaService.empty() {
			gaWidget, err := internal.NewGaWidget(gaService.Keyfile, gaService.ViewID)
			if err != nil {
				internal.DisplayError(tui, err)
			}
			project.WithGa(gaWidget)
		}

		gscService := p.Services.GoogleSearchConsole
		if !gscService.empty() {
			gscWidget, err := internal.NewGscWidget(gscService.Keyfile, gscService.Address)
			if err != nil {
				internal.DisplayError(tui, err)
			}
			project.WithGoogleSearchConsole(gscWidget)
		}

		monService := p.Services.Monitor
		if !monService.empty() {
			monWidget, err := internal.NewMonitorWidget(monService.Address)
			if err != nil {
				internal.DisplayError(tui, err)
			}
			project.WithMonitor(monWidget)
		}

		githubService := p.Services.Github
		if !githubService.empty() {
			githubWidget, err := internal.NewGithubWidget(
				githubService.Token,
				githubService.Owner,
				githubService.Repository,
			)
			if err != nil {
				internal.DisplayError(tui, err)
			}
			project.WithGithub(githubWidget)
		}

		project.Render(tui, *debug)
	}
}
